-- Seed the wilayats for the remaining 10 governorates (migration 7 only seeded
-- Muscat's). Each row is (governorate slug, wilayat name_ar, name_en, slug),
-- joined to its parent governorate by slug.
INSERT INTO locations (parent_id, type, name_ar, name_en, slug)
SELECT g.id, 'wilayat', v.name_ar, v.name_en, v.slug
FROM locations g
JOIN (VALUES
    -- ظفار (Dhofar)
    ('dhofar', 'صلالة',                'Salalah',              'salalah'),
    ('dhofar', 'طاقة',                 'Taqah',                'taqah'),
    ('dhofar', 'مرباط',                'Mirbat',               'mirbat'),
    ('dhofar', 'سدح',                  'Sadah',                'sadah'),
    ('dhofar', 'رخيوت',                'Rakhyut',              'rakhyut'),
    ('dhofar', 'ثمريت',                'Thumrait',             'thumrait'),
    ('dhofar', 'ضلكوت',                'Dhalkut',              'dhalkut'),
    ('dhofar', 'مقشن',                 'Muqshin',              'muqshin'),
    ('dhofar', 'شليم وجزر الحلانيات',  'Shalim',               'shalim'),
    ('dhofar', 'المزيونة',             'Al Mazyona',           'al-mazyona'),
    -- مسندم (Musandam)
    ('musandam', 'خصب',                'Khasab',               'khasab'),
    ('musandam', 'بخا',                'Bukha',                'bukha'),
    ('musandam', 'دبا',                'Daba',                 'daba-musandam'),
    ('musandam', 'مدحاء',              'Madha',                'madha'),
    -- البريمي (Al Buraimi)
    ('al-buraimi', 'البريمي',          'Al Buraimi',           'al-buraimi-wilayat'),
    ('al-buraimi', 'محضة',             'Mahdah',               'mahdah'),
    ('al-buraimi', 'السنينة',          'As Sunaynah',          'as-sunaynah'),
    -- الداخلية (Ad Dakhiliyah)
    ('ad-dakhiliyah', 'نزوى',          'Nizwa',                'nizwa'),
    ('ad-dakhiliyah', 'بهلاء',         'Bahla',                'bahla'),
    ('ad-dakhiliyah', 'منح',           'Manah',                'manah'),
    ('ad-dakhiliyah', 'الحمراء',       'Al Hamra',             'al-hamra'),
    ('ad-dakhiliyah', 'أدم',           'Adam',                 'adam'),
    ('ad-dakhiliyah', 'إزكي',          'Izki',                 'izki'),
    ('ad-dakhiliyah', 'سمائل',         'Samail',               'samail'),
    ('ad-dakhiliyah', 'بدبد',          'Bidbid',               'bidbid'),
    -- شمال الباطنة (Al Batinah North)
    ('al-batinah-north', 'صحار',       'Sohar',                'sohar'),
    ('al-batinah-north', 'شناص',       'Shinas',               'shinas'),
    ('al-batinah-north', 'لوى',        'Liwa',                 'liwa'),
    ('al-batinah-north', 'صحم',        'Saham',                'saham'),
    ('al-batinah-north', 'الخابورة',   'Al Khaburah',          'al-khaburah'),
    ('al-batinah-north', 'السويق',     'As Suwaiq',            'as-suwaiq'),
    -- جنوب الباطنة (Al Batinah South)
    ('al-batinah-south', 'الرستاق',    'Ar Rustaq',            'ar-rustaq'),
    ('al-batinah-south', 'العوابي',    'Al Awabi',             'al-awabi'),
    ('al-batinah-south', 'نخل',        'Nakhal',               'nakhal'),
    ('al-batinah-south', 'وادي المعاول','Wadi Al Maawil',      'wadi-al-maawil'),
    ('al-batinah-south', 'بركاء',      'Barka',                'barka'),
    ('al-batinah-south', 'المصنعة',    'Al Musanaah',          'al-musanaah'),
    -- شمال الشرقية (Ash Sharqiyah North)
    ('ash-sharqiyah-north', 'إبراء',   'Ibra',                 'ibra'),
    ('ash-sharqiyah-north', 'المضيبي', 'Al Mudaibi',           'al-mudaibi'),
    ('ash-sharqiyah-north', 'بدية',    'Bidiyah',              'bidiyah'),
    ('ash-sharqiyah-north', 'القابل',  'Al Qabil',             'al-qabil'),
    ('ash-sharqiyah-north', 'وادي بني خالد', 'Wadi Bani Khalid','wadi-bani-khalid'),
    ('ash-sharqiyah-north', 'دماء والطائيين', 'Dima Wa Al Taien','dima-wa-al-taien'),
    -- جنوب الشرقية (Ash Sharqiyah South)
    ('ash-sharqiyah-south', 'صور',     'Sur',                  'sur'),
    ('ash-sharqiyah-south', 'الكامل والوافي', 'Al Kamil Wa Al Wafi', 'al-kamil-wa-al-wafi'),
    ('ash-sharqiyah-south', 'جعلان بني بو حسن', 'Jalan Bani Bu Hasan', 'jalan-bani-bu-hasan'),
    ('ash-sharqiyah-south', 'جعلان بني بو علي', 'Jalan Bani Bu Ali', 'jalan-bani-bu-ali'),
    ('ash-sharqiyah-south', 'مصيرة',   'Masirah',              'masirah'),
    -- الظاهرة (Adh Dhahirah)
    ('adh-dhahirah', 'عبري',           'Ibri',                 'ibri'),
    ('adh-dhahirah', 'ينقل',           'Yanqul',               'yanqul'),
    ('adh-dhahirah', 'ضنك',            'Dank',                 'dank'),
    -- الوسطى (Al Wusta)
    ('al-wusta', 'هيماء',              'Haima',                'haima'),
    ('al-wusta', 'محوت',               'Mahout',               'mahout'),
    ('al-wusta', 'الدقم',              'Ad Duqm',              'ad-duqm'),
    ('al-wusta', 'الجازر',             'Al Jazer',             'al-jazer')
) AS v(gov_slug, name_ar, name_en, slug) ON g.slug = v.gov_slug
WHERE g.type = 'governorate';
