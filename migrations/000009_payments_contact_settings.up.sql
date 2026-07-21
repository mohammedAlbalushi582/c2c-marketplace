-- Paid-listing fees (Thawani-ready), the راسلنا contact inbox, and admin-editable
-- app settings (fee tiers + listing duration). No listings-table change is needed:
-- `expires_at`/`published_at` and the `draft` status already exist from migration 5.

-- ---- payments ----
-- One row per fee charge. Provider-agnostic: `provider`/`provider_ref`/`provider_payload`
-- hold whatever the gateway (stub now, Thawani later) returns.
CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'canceled');

CREATE TABLE payments (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id       bigint REFERENCES listings(id) ON DELETE SET NULL,
    purpose          text NOT NULL DEFAULT 'listing_fee',   -- listing_fee | renewal
    amount           numeric(12,3) NOT NULL,                -- OMR, 3 decimals (baisa)
    currency         char(3) NOT NULL DEFAULT 'OMR',
    status           payment_status NOT NULL DEFAULT 'pending',
    provider         text NOT NULL DEFAULT 'stub',          -- stub | thawani
    provider_ref     text,                                  -- gateway session/checkout id
    provider_payload jsonb,                                 -- raw gateway response (audit)
    paid_at          timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_user    ON payments(user_id);
CREATE INDEX idx_payments_listing ON payments(listing_id);
CREATE INDEX idx_payments_ref     ON payments(provider_ref);
CREATE TRIGGER trg_payments_updated BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---- contact messages (راسلنا) ----
CREATE TABLE contact_messages (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       text NOT NULL,
    phone      text,
    email      text,
    message    text NOT NULL,
    is_read    boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_contact_unread ON contact_messages(is_read, created_at DESC);

-- ---- app settings (admin-editable key/value) ----
CREATE TABLE app_settings (
    key        text PRIMARY KEY,
    value      text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO app_settings (key, value) VALUES
    ('listing_fee_tier2', '2.000'),      -- fee for the user's 2nd live listing
    ('listing_fee_tier3_plus', '5.000'), -- fee for the 3rd and every listing after
    ('listing_duration_days', '30'),     -- how long an approved listing stays active
    ('payment_provider', 'stub');        -- informational; real selection is env-driven
