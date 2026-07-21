package payment

import (
	"context"
	"fmt"
	"strings"
)

// Stub is the development/demo provider. It doesn't move money: CreateCheckout
// sends the buyer to the app's own /pay/{id} page, and FetchStatus always
// reports "paid" — so you can click through the entire paid-listing flow with
// no gateway account. Selected when PAYMENT_PROVIDER=stub (the default).
type Stub struct {
	frontendBaseURL string
}

// NewStub builds the stub provider. frontendBaseURL is where the simulated
// checkout page lives (the Next.js app).
func NewStub(frontendBaseURL string) *Stub {
	return &Stub{frontendBaseURL: strings.TrimRight(frontendBaseURL, "/")}
}

func (s *Stub) Name() string { return "stub" }

func (s *Stub) CreateCheckout(_ context.Context, req CheckoutRequest) (*Checkout, error) {
	return &Checkout{
		ProviderRef: fmt.Sprintf("stub_%d", req.PaymentID),
		CheckoutURL: fmt.Sprintf("%s/pay/%d", s.frontendBaseURL, req.PaymentID),
	}, nil
}

// FetchStatus simulates a successful payment — the stub /pay page triggers the
// verify call only after the user clicks "pay", so returning true here mirrors
// a completed gateway checkout.
func (s *Stub) FetchStatus(_ context.Context, _ string) (*Status, error) {
	return &Status{Paid: true}, nil
}
