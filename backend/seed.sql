-- Development data only (never run in production). Accounts here have no Auth0 identity: to log in
-- as one, set auth0_sub to your own test user's subject, e.g.
--   UPDATE users SET auth0_sub = 'auth0|YOUR_SUB' WHERE email = 'agency@example.com';
INSERT INTO users (id, auth0_sub, email, role) VALUES
  ('00000000-0000-0000-0000-0000000000a1', 'seed|agency',  'agency@example.com',  'agency'),
  ('00000000-0000-0000-0000-0000000000b1', 'seed|hr',      'hr@example.com',      'hr'),
  ('00000000-0000-0000-0000-0000000000c1', 'seed|newcomer','newcomer@example.com','newcomer')
ON CONFLICT DO NOTHING;
INSERT INTO agencies (id, name, city, country, verified_at) VALUES
  ('10000000-0000-0000-0000-000000000001', 'Casa Bogotá Demo', 'Bogotá', 'CO', now()) ON CONFLICT DO NOTHING;
INSERT INTO agency_profiles (user_id, agency_id, full_name) VALUES
  ('00000000-0000-0000-0000-0000000000a1', '10000000-0000-0000-0000-000000000001', 'Ana Agent') ON CONFLICT DO NOTHING;
INSERT INTO companies (id, name, industry, city, country) VALUES
  ('20000000-0000-0000-0000-000000000001', 'Demo Employer', 'IT', 'Bogotá', 'CO') ON CONFLICT DO NOTHING;
INSERT INTO hr_profiles (user_id, company_id, full_name) VALUES
  ('00000000-0000-0000-0000-0000000000b1', '20000000-0000-0000-0000-000000000001', 'Rita HR') ON CONFLICT DO NOTHING;
INSERT INTO newcomers (id, user_id, first_name, city_of_contract, city_key, budget_max, bedrooms, furnished, stage, allow_contact) VALUES
  ('30000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-0000000000c1', 'Maria', 'Bogotá', 'bogota', 3000000, 2, true, 'searching', true)
  ON CONFLICT DO NOTHING;
INSERT INTO newcomer_contacts (newcomer_id, full_name, email, phone) VALUES
  ('30000000-0000-0000-0000-000000000001', 'Maria Lopez', 'newcomer@example.com', '+57 300 000 0000') ON CONFLICT DO NOTHING;
INSERT INTO properties (agency_id, title, price, type, bedrooms, furnished, city, city_key, district, address) VALUES
  ('10000000-0000-0000-0000-000000000001', 'Bright 2-bedroom in Chapinero', 2800000, 'apartment', 2, true, 'Bogotá', 'bogota', 'Chapinero', 'Calle 60 # 10-20');
