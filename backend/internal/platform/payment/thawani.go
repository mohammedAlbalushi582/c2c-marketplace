package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// Thawani integrates Thawani Pay's e-commerce checkout (https://thawani.om),
// Oman's payment gateway. This is the "drop-in space": the wiring is complete,
// so enabling it is just configuration —
//
//	PAYMENT_PROVIDER=thawani
//	THAWANI_BASE_URL=https://checkout.thawani.om/api/v1   (UAT: https://uatcheckout.thawani.om/api/v1)
//	THAWANI_SECRET_KEY=<your secret key>
//	THAWANI_PUBLISHABLE_KEY=<your publishable key>
//
// Thawani works in baisa (1 OMR = 1000 baisa) and authenticates server calls
// with the `thawani-api-key` header. The buyer is redirected to the hosted
// checkout page and returns to SuccessURL/CancelURL; FetchStatus then confirms
// the session's payment_status before the listing is released.
//
// NOTE: verify field names/limits against the current Thawani API docs when you
// go live — this reflects the documented v1 shape but has not been run against
// a live merchant account here.
type Thawani struct {
	baseURL        string // e.g. https://uatcheckout.thawani.om/api/v1
	secretKey      string // thawani-api-key header (server-side)
	publishableKey string // used to build the hosted-checkout redirect URL
	http           *http.Client
}

// NewThawani builds the provider. Returns an error if required keys are missing
// so misconfiguration fails fast at startup rather than at first payment.
func NewThawani(baseURL, secretKey, publishableKey string) (*Thawani, error) {
	if secretKey == "" || publishableKey == "" {
		return nil, errors.New("thawani: THAWANI_SECRET_KEY and THAWANI_PUBLISHABLE_KEY are required")
	}
	return &Thawani{
		baseURL:        strings.TrimRight(baseURL, "/"),
		secretKey:      secretKey,
		publishableKey: publishableKey,
		http:           &http.Client{Timeout: 20 * time.Second},
	}, nil
}

func (t *Thawani) Name() string { return "thawani" }

// hostedCheckoutBase turns the API base (…/api/v1) into the pay host used for
// the buyer-facing redirect (…/pay/{session_id}?key=…).
func (t *Thawani) hostedCheckoutBase() string {
	return strings.TrimSuffix(t.baseURL, "/api/v1")
}

func (t *Thawani) CreateCheckout(ctx context.Context, req CheckoutRequest) (*Checkout, error) {
	body := map[string]any{
		"client_reference_id": fmt.Sprintf("%d", req.PaymentID),
		"mode":                "payment",
		"products": []map[string]any{{
			"name":        req.Description,
			"quantity":    1,
			"unit_amount": toBaisa(req.Amount), // integer baisa
		}},
		"success_url": req.SuccessURL,
		"cancel_url":  req.CancelURL,
		"metadata":    req.Metadata,
	}
	raw, err := t.do(ctx, http.MethodPost, "/checkout/session", body)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Success bool `json:"success"`
		Data    struct {
			SessionID string `json:"session_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("thawani: decode session: %w", err)
	}
	if !parsed.Success || parsed.Data.SessionID == "" {
		return nil, fmt.Errorf("thawani: checkout not created: %s", raw)
	}
	return &Checkout{
		ProviderRef: parsed.Data.SessionID,
		CheckoutURL: fmt.Sprintf("%s/pay/%s?key=%s", t.hostedCheckoutBase(), parsed.Data.SessionID, t.publishableKey),
		RawPayload:  raw,
	}, nil
}

func (t *Thawani) FetchStatus(ctx context.Context, providerRef string) (*Status, error) {
	raw, err := t.do(ctx, http.MethodGet, "/checkout/session/"+providerRef, nil)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Data struct {
			PaymentStatus string `json:"payment_status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("thawani: decode status: %w", err)
	}
	return &Status{Paid: parsed.Data.PaymentStatus == "paid", RawPayload: raw}, nil
}

// do performs an authenticated Thawani API call and returns the raw body.
func (t *Thawani) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewReader(b)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, t.baseURL+path, buf)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("thawani-api-key", t.secretKey)

	resp, err := t.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("thawani: %s %s -> %d: %s", method, path, resp.StatusCode, raw)
	}
	return raw, nil
}

// toBaisa converts OMR (major units) to integer baisa (1 OMR = 1000 baisa).
func toBaisa(omr float64) int64 {
	return int64(math.Round(omr * 1000))
}
