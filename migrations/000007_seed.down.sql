DELETE FROM locations WHERE slug IN (
    'muscat-wilayat','muttrah','bawshar','as-seeb','al-amarat','quriyat',
    'muscat','dhofar','musandam','al-buraimi','ad-dakhiliyah',
    'al-batinah-north','al-batinah-south','ash-sharqiyah-north',
    'ash-sharqiyah-south','adh-dhahirah','al-wusta'
);
DELETE FROM category_fields WHERE category_id IN (
    SELECT id FROM categories WHERE slug IN ('real-estate','website-templates')
);
DELETE FROM categories WHERE slug IN ('real-estate','website-templates');
