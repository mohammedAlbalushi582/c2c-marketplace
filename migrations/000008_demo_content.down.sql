DELETE FROM listings WHERE slug IN (
    'land-seeb','house-bawshar','apt-muttrah','land-amarat','tpl-store-next','tpl-realestate'
);
DELETE FROM users WHERE email IN ('admin@amjad.om','seller@amjad.om');
