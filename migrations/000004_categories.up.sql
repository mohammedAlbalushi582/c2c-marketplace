CREATE TYPE field_type AS ENUM
    ('text','textarea','number','select','multiselect','boolean','date','url');

-- Self-referencing category tree; admin can add subcategories at any depth.
CREATE TABLE categories (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id     bigint REFERENCES categories(id) ON DELETE RESTRICT,
    slug          text NOT NULL UNIQUE,
    name_ar       text NOT NULL,
    name_en       text,
    icon          text,
    display_order int  NOT NULL DEFAULT 0,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_categories_parent ON categories(parent_id);
CREATE TRIGGER trg_categories_updated BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Dynamic per-category attribute definitions.
-- e.g. land: area_size / plot_number; template: tech_stack / demo_link.
CREATE TABLE category_fields (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id   bigint NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    field_key     text   NOT NULL,          -- machine name, e.g. 'area_size'
    label_ar      text   NOT NULL,
    label_en      text,
    field_type    field_type NOT NULL,
    unit          text,                      -- e.g. 'm²'
    options       jsonb,                     -- choices for select / multiselect
    is_required   boolean NOT NULL DEFAULT false,
    is_filterable boolean NOT NULL DEFAULT false,
    display_order int NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (category_id, field_key)
);
CREATE INDEX idx_category_fields_category ON category_fields(category_id);
CREATE TRIGGER trg_category_fields_updated BEFORE UPDATE ON category_fields
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
