// Package payment abstracts the payment gateway behind an interface so the
// dev "stub" provider can be swapped for Thawani (or any gateway) without
// touching the usecase layer — mirroring the storage.Storage seam.
//
// Flow shared by every provider:
//  1. CreateCheckout — hand the gateway the amount + return URLs, get back a
//     CheckoutURL to redirect the buyer to and a ProviderRef to remember.
//  2. FetchStatus — after the buyer returns, ask the gateway whether the
//     ProviderRef was actually paid. The listing is only released on 'paid'.
package payment

import "context"

// CheckoutRequest is what the app hands the gateway to start a payment.
type CheckoutRequest struct {
	PaymentID   int64             // our payments.id — used as client reference
	Amount      float64           // in OMR (major units)
	Currency    string            // "OMR"
	Description string            // human label, e.g. "رسوم نشر إعلان"
	SuccessURL  string            // where the gateway returns the buyer on success
	CancelURL   string            // where the gateway returns the buyer on cancel
	Metadata    map[string]string // echoed back where the gateway supports it
}

// Checkout is the gateway's response to CreateCheckout.
type Checkout struct {
	ProviderRef string // gateway session/checkout id (stored on the payment row)
	CheckoutURL string // URL to redirect the buyer to
	RawPayload  []byte // raw gateway response, kept for audit
}

// Status is the normalized result of querying a checkout after the buyer returns.
type Status struct {
	Paid       bool
	RawPayload []byte
}

// Provider is the payment-gateway seam. Swap the concrete type in main.go
// (driven by PAYMENT_PROVIDER) with no change to callers.
type Provider interface {
	// Name is the value stored in payments.provider (e.g. "stub", "thawani").
	Name() string
	// CreateCheckout starts a payment and returns where to send the buyer.
	CreateCheckout(ctx context.Context, req CheckoutRequest) (*Checkout, error)
	// FetchStatus reports whether the checkout identified by providerRef is paid.
	FetchStatus(ctx context.Context, providerRef string) (*Status, error)
}
