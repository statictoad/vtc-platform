-- internal/client/seeds/clients-data.sql
INSERT INTO client.clients (id, clerk_user_id, first_name, last_name, email, can_pay_later, created_at, updated_at)
VALUES
  ('019600cf-1000-7000-8000-000000000001', 'user_seed_001', 'Jean',  'Dupont', 'jean.dupont@example.com',  false, NOW(), NOW()),
  ('019600cf-2000-7000-8000-000000000002', 'user_seed_002', 'Marie', 'Martin', 'marie.martin@example.com', false, NOW(), NOW())
ON CONFLICT DO NOTHING;