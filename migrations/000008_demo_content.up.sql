-- Demo users and active listings so the site has visible content out of the box.
-- Passwords are all "password123" (bcrypt hashed). Remove in production.

INSERT INTO users (email, username, password_hash, full_name, phone, whatsapp_number, role) VALUES
    ('admin@amjad.om',  'admin',  '$2a$10$iU/OOvImcq1iN71h/37M/uc4gMGUbhOzZWDjG83DHCGvsJRqcsSCK', 'مدير النظام',  '96890000001', '96890000001', 'admin'),
    ('seller@amjad.om', 'seller', '$2a$10$iU/OOvImcq1iN71h/37M/uc4gMGUbhOzZWDjG83DHCGvsJRqcsSCK', 'أحمد البلوشي', '96890123456', '96890123456', 'user');

-- Listings (all approved/active).
INSERT INTO listings
    (user_id, category_id, location_id, title, slug, description, price, currency, price_type,
     contact_phone, whatsapp_number, status, is_featured, published_at)
SELECT u.id, c.id, loc.id, v.title, v.slug, v.descr, v.price, 'OMR', v.ptype::price_type,
       '96890123456', '96890123456', 'active', v.featured, now()
FROM (VALUES
    ('أرض سكنية في السيب',          'land-seeb',        'أرض سكنية ممتازة بموقع مميز قرب جميع الخدمات والمرافق العامة.', 35000::numeric, 'negotiable', 'real-estate',       'as-seeb',   true),
    ('منزل واسع للبيع في بوشر',      'house-bawshar',    'منزل حديث مكوّن من طابقين مع حديقة ومواقف سيارات.',            85000::numeric, 'fixed',      'real-estate',       'bawshar',   false),
    ('شقة فاخرة في مطرح',           'apt-muttrah',      'شقة بإطلالة بحرية رائعة قريبة من الكورنيش والأسواق.',          45000::numeric, 'negotiable', 'real-estate',       'muttrah',   false),
    ('أرض تجارية في العامرات',       'land-amarat',      'أرض تجارية على شارع رئيسي، مناسبة للمشاريع الاستثمارية.',       120000::numeric,'on_request', 'real-estate',       'al-amarat', true),
    ('قالب متجر إلكتروني Next.js',   'tpl-store-next',   'قالب متجر إلكتروني متكامل مبني بـ Next.js وجاهز للنشر.',        150::numeric,   'fixed',      'website-templates', NULL,        false),
    ('قالب موقع شركة عقارية',        'tpl-realestate',   'قالب احترافي لعرض العقارات مع لوحة تحكم كاملة.',               90::numeric,    'negotiable', 'website-templates', NULL,        false)
) AS v(title, slug, descr, price, ptype, cat_slug, loc_slug, featured)
JOIN categories c ON c.slug = v.cat_slug
JOIN users u ON u.email = 'seller@amjad.om'
LEFT JOIN locations loc ON loc.slug = v.loc_slug;

-- Real-estate attributes.
INSERT INTO listing_attributes (listing_id, field_id, value_number)
SELECT l.id, f.id, v.num
FROM (VALUES
    ('land-seeb', 'area_size', 600::numeric),
    ('house-bawshar', 'area_size', 400::numeric),
    ('house-bawshar', 'bedrooms', 5::numeric),
    ('apt-muttrah', 'area_size', 180::numeric),
    ('apt-muttrah', 'bedrooms', 3::numeric),
    ('land-amarat', 'area_size', 1000::numeric)
) AS v(slug, key, num)
JOIN listings l ON l.slug = v.slug
JOIN category_fields f ON f.field_key = v.key AND f.category_id = l.category_id;

INSERT INTO listing_attributes (listing_id, field_id, value_text)
SELECT l.id, f.id, v.txt
FROM (VALUES
    ('land-seeb', 'property_type', 'land'),
    ('house-bawshar', 'property_type', 'house'),
    ('apt-muttrah', 'property_type', 'apartment'),
    ('land-amarat', 'property_type', 'commercial'),
    ('tpl-store-next', 'demo_link', 'https://demo.example.com/store'),
    ('tpl-realestate', 'demo_link', 'https://demo.example.com/realestate'),
    ('tpl-store-next', 'license_type', 'single'),
    ('tpl-realestate', 'license_type', 'extended')
) AS v(slug, key, txt)
JOIN listings l ON l.slug = v.slug
JOIN category_fields f ON f.field_key = v.key AND f.category_id = l.category_id;

INSERT INTO listing_attributes (listing_id, field_id, value_json)
SELECT l.id, f.id, v.js::jsonb
FROM (VALUES
    ('tpl-store-next', 'tech_stack', '["nextjs","react"]'),
    ('tpl-realestate', 'tech_stack', '["html_css","wordpress"]')
) AS v(slug, key, js)
JOIN listings l ON l.slug = v.slug
JOIN category_fields f ON f.field_key = v.key AND f.category_id = l.category_id;

INSERT INTO listing_attributes (listing_id, field_id, value_boolean)
SELECT l.id, f.id, true
FROM listings l
JOIN category_fields f ON f.field_key = 'responsive' AND f.category_id = l.category_id
WHERE l.slug IN ('tpl-store-next', 'tpl-realestate');
