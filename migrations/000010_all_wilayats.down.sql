-- Remove the wilayats added here (everything except Muscat's, seeded in 7).
DELETE FROM locations
WHERE type = 'wilayat'
  AND parent_id IN (SELECT id FROM locations WHERE type = 'governorate' AND slug <> 'muscat');
