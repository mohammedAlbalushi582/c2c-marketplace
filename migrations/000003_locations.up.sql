-- Oman administrative hierarchy: governorate (محافظة) -> wilayat (ولاية).
CREATE TYPE location_type AS ENUM ('governorate', 'wilayat');

CREATE TABLE locations (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id  bigint REFERENCES locations(id) ON DELETE CASCADE,
    type       location_type NOT NULL,
    name_ar    text NOT NULL,
    name_en    text,
    slug       text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_locations_parent ON locations(parent_id);
