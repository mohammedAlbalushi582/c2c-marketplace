-- ---------------------------------------------------------------------------
-- Seed data: root categories + their dynamic fields, and Oman locations.
-- Idempotent-ish for a fresh DB; safe to run once after schema migrations.
-- ---------------------------------------------------------------------------

-- Root categories -----------------------------------------------------------
INSERT INTO categories (slug, name_ar, name_en, icon, display_order) VALUES
    ('real-estate',      'بيع الاراضي والمنازل والمواقع', 'Real Estate & Land', 'home',  1),
    ('website-templates','بيع قوالب المواقع',            'Website Templates',  'code',  2);

-- Real-estate custom fields -------------------------------------------------
INSERT INTO category_fields
    (category_id, field_key, label_ar, label_en, field_type, unit, options, is_required, is_filterable, display_order)
SELECT c.id, v.field_key, v.label_ar, v.label_en, v.field_type::field_type, v.unit, v.options::jsonb,
       v.is_required, v.is_filterable, v.display_order
FROM categories c
CROSS JOIN (VALUES
    ('area_size',     'المساحة',      'Area size',      'number', 'm²', NULL,
        true,  true,  1),
    ('plot_number',   'رقم القطعة',   'Plot number',    'text',   NULL, NULL,
        false, false, 2),
    ('property_type', 'نوع العقار',   'Property type',  'select', NULL,
        '[{"value":"land","label_ar":"أرض","label_en":"Land"},
          {"value":"house","label_ar":"منزل","label_en":"House"},
          {"value":"apartment","label_ar":"شقة","label_en":"Apartment"},
          {"value":"commercial","label_ar":"تجاري","label_en":"Commercial"}]',
        true,  true,  3),
    ('bedrooms',      'غرف النوم',    'Bedrooms',       'number', NULL, NULL,
        false, true,  4)
) AS v(field_key, label_ar, label_en, field_type, unit, options, is_required, is_filterable, display_order)
WHERE c.slug = 'real-estate';

-- Website-template custom fields --------------------------------------------
INSERT INTO category_fields
    (category_id, field_key, label_ar, label_en, field_type, unit, options, is_required, is_filterable, display_order)
SELECT c.id, v.field_key, v.label_ar, v.label_en, v.field_type::field_type, v.unit, v.options::jsonb,
       v.is_required, v.is_filterable, v.display_order
FROM categories c
CROSS JOIN (VALUES
    ('tech_stack',   'التقنيات',       'Tech stack',   'multiselect', NULL,
        '[{"value":"html_css","label_ar":"HTML/CSS","label_en":"HTML/CSS"},
          {"value":"react","label_ar":"React","label_en":"React"},
          {"value":"nextjs","label_ar":"Next.js","label_en":"Next.js"},
          {"value":"vue","label_ar":"Vue","label_en":"Vue"},
          {"value":"wordpress","label_ar":"WordPress","label_en":"WordPress"},
          {"value":"laravel","label_ar":"Laravel","label_en":"Laravel"}]',
        false, true,  1),
    ('demo_link',    'رابط المعاينة',  'Demo link',    'url',     NULL, NULL,
        false, false, 2),
    ('responsive',   'متجاوب',         'Responsive',   'boolean', NULL, NULL,
        false, true,  3),
    ('license_type', 'نوع الترخيص',    'License type', 'select',  NULL,
        '[{"value":"single","label_ar":"ترخيص فردي","label_en":"Single"},
          {"value":"extended","label_ar":"ترخيص موسع","label_en":"Extended"}]',
        false, true,  4)
) AS v(field_key, label_ar, label_en, field_type, unit, options, is_required, is_filterable, display_order)
WHERE c.slug = 'website-templates';

-- Oman governorates ----------------------------------------------------------
INSERT INTO locations (type, name_ar, name_en, slug) VALUES
    ('governorate', 'مسقط',          'Muscat',              'muscat'),
    ('governorate', 'ظفار',          'Dhofar',              'dhofar'),
    ('governorate', 'مسندم',         'Musandam',            'musandam'),
    ('governorate', 'البريمي',       'Al Buraimi',          'al-buraimi'),
    ('governorate', 'الداخلية',      'Ad Dakhiliyah',       'ad-dakhiliyah'),
    ('governorate', 'شمال الباطنة',  'Al Batinah North',    'al-batinah-north'),
    ('governorate', 'جنوب الباطنة',  'Al Batinah South',    'al-batinah-south'),
    ('governorate', 'شمال الشرقية',  'Ash Sharqiyah North', 'ash-sharqiyah-north'),
    ('governorate', 'جنوب الشرقية',  'Ash Sharqiyah South', 'ash-sharqiyah-south'),
    ('governorate', 'الظاهرة',       'Adh Dhahirah',        'adh-dhahirah'),
    ('governorate', 'الوسطى',        'Al Wusta',            'al-wusta');

-- Muscat wilayats (other governorates' wilayats can be added via admin later).
INSERT INTO locations (parent_id, type, name_ar, name_en, slug)
SELECT g.id, 'wilayat', v.name_ar, v.name_en, v.slug
FROM locations g
CROSS JOIN (VALUES
    ('مسقط',    'Muscat',    'muscat-wilayat'),
    ('مطرح',    'Muttrah',   'muttrah'),
    ('بوشر',    'Bawshar',   'bawshar'),
    ('السيب',   'As Seeb',   'as-seeb'),
    ('العامرات','Al Amarat', 'al-amarat'),
    ('قريات',   'Quriyat',   'quriyat')
) AS v(name_ar, name_en, slug)
WHERE g.slug = 'muscat';
