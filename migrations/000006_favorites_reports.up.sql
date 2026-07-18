CREATE TABLE favorites (
    user_id    bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id bigint NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, listing_id)
);
CREATE INDEX idx_favorites_listing ON favorites(listing_id);

CREATE TYPE report_status AS ENUM ('open','reviewed','dismissed');

CREATE TABLE reports (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    listing_id       bigint NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    reporter_user_id bigint REFERENCES users(id) ON DELETE SET NULL,
    reason           text NOT NULL,
    details          text,
    status           report_status NOT NULL DEFAULT 'open',
    created_at       timestamptz NOT NULL DEFAULT now(),
    reviewed_at      timestamptz,
    reviewed_by      bigint REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX idx_reports_listing ON reports(listing_id);
CREATE INDEX idx_reports_status  ON reports(status);
