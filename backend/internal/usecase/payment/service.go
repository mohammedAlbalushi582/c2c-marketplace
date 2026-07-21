// Package payment (usecase) prices listing fees, opens gateway checkouts, and
// releases a listing into moderation once its fee is confirmed paid. It talks to
// the gateway through the platform payment.Provider seam, so the same logic runs
// against the dev stub or Thawani.
package payment

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/alamjad/marketplace/internal/domain"
	paygw "github.com/alamjad/marketplace/internal/platform/payment"
	"github.com/alamjad/marketplace/internal/platform/pgutil"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
)

// Settings keys + fallbacks (used when a row is missing or unpar. These mirror
// the seed in migration 000009.
const (
	keyFeeTier2     = "listing_fee_tier2"
	keyFeeTier3Plus = "listing_fee_tier3_plus"
	defaultFeeTier2 = 2.0
	defaultFeeTier3 = 5.0
	currencyOMR     = "OMR"
	purposeListing  = "listing_fee"
	purposeRenewal  = "renewal"
)

// Store is the persistence seam: generated queries plus transactions.
type Store interface {
	sqlc.Querier
	WithTx(ctx context.Context, fn func(*sqlc.Queries) error) error
}

type Service struct {
	store           Store
	provider        paygw.Provider
	frontendBaseURL string
}

func NewService(store Store, provider paygw.Provider, frontendBaseURL string) *Service {
	return &Service{store: store, provider: provider, frontendBaseURL: strings.TrimRight(frontendBaseURL, "/")}
}

// Quote is the fee owed for a user's next listing.
type Quote struct {
	Fee       float64 `json:"fee"`
	Currency  string  `json:"currency"`
	Tier      int     `json:"tier"`       // 1 = first (free), 2, 3
	LiveCount int64   `json:"live_count"` // current pending+active listings
	Free      bool    `json:"free"`
}

// QuoteListingFee prices the next listing from how many live listings the user
// already has: 0 → free, 1 → tier2, 2+ → tier3_plus. Amounts are admin-editable.
func (s *Service) QuoteListingFee(ctx context.Context, userID int64) (*Quote, error) {
	n, err := s.store.CountLiveListingsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	q := &Quote{Currency: currencyOMR, LiveCount: n}
	switch {
	case n <= 0:
		q.Tier, q.Fee, q.Free = 1, 0, true
	case n == 1:
		q.Tier, q.Fee = 2, s.settingFloat(ctx, keyFeeTier2, defaultFeeTier2)
	default:
		q.Tier, q.Fee = 3, s.settingFloat(ctx, keyFeeTier3Plus, defaultFeeTier3)
	}
	q.Free = q.Fee <= 0
	return q, nil
}

// Start opens a checkout for a listing fee (or renewal) and returns the payment
// row plus the URL to send the buyer to.
type Start struct {
	Payment     sqlc.Payment
	CheckoutURL string
}

// StartListingPayment creates a pending payment for listingID and opens a
// gateway checkout. purpose is purposeListing or purposeRenewal.
func (s *Service) StartListingPayment(ctx context.Context, userID, listingID int64, amount float64, purpose string) (*Start, error) {
	lid := listingID
	p, err := s.store.CreatePayment(ctx, sqlc.CreatePaymentParams{
		UserID:    userID,
		ListingID: &lid,
		Purpose:   purpose,
		Amount:    pgutil.NumericFromFloatPtr(&amount),
		Currency:  currencyOMR,
		Provider:  s.provider.Name(),
	})
	if err != nil {
		return nil, err
	}

	co, err := s.provider.CreateCheckout(ctx, paygw.CheckoutRequest{
		PaymentID:   p.ID,
		Amount:      amount,
		Currency:    currencyOMR,
		Description: "رسوم نشر إعلان",
		SuccessURL:  s.returnURL(p.ID, "success"),
		CancelURL:   s.returnURL(p.ID, "cancel"),
		Metadata:    map[string]string{"listing_id": strconv.FormatInt(listingID, 10)},
	})
	if err != nil {
		// Leave the payment pending; the buyer can retry. Surface the failure.
		return nil, err
	}

	p, err = s.store.SetPaymentCheckout(ctx, sqlc.SetPaymentCheckoutParams{
		ID:              p.ID,
		ProviderRef:     &co.ProviderRef,
		ProviderPayload: co.RawPayload,
	})
	if err != nil {
		return nil, err
	}
	return &Start{Payment: p, CheckoutURL: co.CheckoutURL}, nil
}

// VerifyResult reports the outcome of confirming a checkout.
type VerifyResult struct {
	Status  sqlc.PaymentStatus `json:"status"`
	Paid    bool               `json:"paid"`
	Payment sqlc.Payment       `json:"-"`
}

// Verify asks the gateway whether the payment's checkout is paid and, if so,
// marks it paid and moves its listing into moderation. Idempotent and
// ownership-checked. Called when the buyer returns from checkout.
func (s *Service) Verify(ctx context.Context, userID, paymentID int64) (*VerifyResult, error) {
	p, err := s.store.GetPayment(ctx, paymentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if p.Status == sqlc.PaymentStatusPaid {
		return &VerifyResult{Status: p.Status, Paid: true, Payment: p}, nil
	}
	if p.Status != sqlc.PaymentStatusPending {
		return &VerifyResult{Status: p.Status, Paid: false, Payment: p}, nil
	}

	ref := ""
	if p.ProviderRef != nil {
		ref = *p.ProviderRef
	}
	st, err := s.provider.FetchStatus(ctx, ref)
	if err != nil {
		return nil, err
	}
	if !st.Paid {
		return &VerifyResult{Status: sqlc.PaymentStatusPending, Paid: false, Payment: p}, nil
	}

	// Mark paid and release the listing atomically.
	err = s.store.WithTx(ctx, func(q *sqlc.Queries) error {
		paid, err := q.MarkPaymentPaid(ctx, sqlc.MarkPaymentPaidParams{ID: p.ID, ProviderPayload: st.RawPayload})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err // ErrNoRows => already paid by a concurrent verify; fine
		}
		if err == nil {
			p = paid
		}
		if p.ListingID != nil {
			// draft/expired -> pending; a no-op (ErrNoRows) if already queued.
			if _, err := q.SubmitListingForReview(ctx, *p.ListingID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &VerifyResult{Status: sqlc.PaymentStatusPaid, Paid: true, Payment: p}, nil
}

// Get returns a payment the user owns (for status polling).
func (s *Service) Get(ctx context.Context, userID, paymentID int64) (sqlc.Payment, error) {
	p, err := s.store.GetPayment(ctx, paymentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Payment{}, domain.ErrNotFound
	} else if err != nil {
		return sqlc.Payment{}, err
	}
	if p.UserID != userID {
		return sqlc.Payment{}, domain.ErrForbidden
	}
	return p, nil
}

func (s *Service) returnURL(paymentID int64, outcome string) string {
	return s.frontendBaseURL + "/pay/" + strconv.FormatInt(paymentID, 10) + "?status=" + outcome
}

func (s *Service) settingFloat(ctx context.Context, key string, def float64) float64 {
	v, err := s.store.GetSetting(ctx, key)
	if err != nil {
		return def
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return def
	}
	return f
}
