-- internal/fleet/seeds/vehicles-data.sql
INSERT INTO fleet.vehicles (id, number_plate, brand, model, year, active, created_at, updated_at)
VALUES
  ('019600cf-3000-7000-8000-000000000003', 'AB-123-CD', 'Mercedes-Benz', 'Classe E', 2022, true, NOW(), NOW()),
  ('019600cf-4000-7000-8000-000000000004', 'EF-456-GH', 'BMW',           'Série 5',  2023, true, NOW(), NOW())
ON CONFLICT DO NOTHING;