CREATE TYPE listing_status AS ENUM ('draft','pending','active','sold','expired','rejected');
CREATE TYPE price_type     AS ENUM ('fixed','negotiable','on_request');

CREATE TABLE listings (
    id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id     bigint NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    location_id     bigint REFERENCES locations(id) ON DELETE SET NULL,
    title           text NOT NULL,
    slug            text,
    description     text NOT NULL,
    price           numeric(14,3),           -- OMR: 3 decimal places (baisa)
    currency        char(3) NOT NULL DEFAULT 'OMR',
    price_type      price_type NOT NULL DEFAULT 'fixed',
    contact_phone   text,
    whatsapp_number text,
    status          listing_status NOT NULL DEFAULT 'pending',  -- admin approval before 'active'
    is_featured     boolean NOT NULL DEFAULT false,
    featured_until  timestamptz,
    views_count     bigint NOT NULL DEFAULT 0,
    published_at    timestamptz,
    expires_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz              -- soft delete
);
CREATE INDEX idx_listings_category ON listings(category_id);
CREATE INDEX idx_listings_location ON listings(location_id);
CREATE INDEX idx_listings_user     ON listings(user_id);
CREATE INDEX idx_listings_status   ON listings(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_listings_active_feed
    ON listings(published_at DESC) WHERE status = 'active' AND deleted_at IS NULL;
CREATE TRIGGER trg_listings_updated BEFORE UPDATE ON listings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE listing_images (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    listing_id    bigint NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    storage_key   text NOT NULL,             -- path/key; public URL built by storage layer
    alt_text      text,
    is_primary    boolean NOT NULL DEFAULT false,
    display_order int NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_listing_images_listing ON listing_images(listing_id);

-- EAV: category-specific field values. Typed columns keep filtering indexable.
CREATE TABLE listing_attributes (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    listing_id    bigint NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    field_id      bigint NOT NULL REFERENCES category_fields(id) ON DELETE CASCADE,
    value_text    text,
    value_number  numeric(20,4),
    value_boolean boolean,
    value_date    date,
    value_json    jsonb,                      -- multiselect values
    UNIQUE (listing_id, field_id)
);
CREATE INDEX idx_lattr_field_number ON listing_attributes(field_id, value_number);
CREATE INDEX idx_lattr_field_text   ON listing_attributes(field_id, value_text);
