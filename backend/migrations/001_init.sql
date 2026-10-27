-- Bogoboom schema (PostgreSQL 15+). Applied in order by `go run ./cmd/migrate`.
-- Auth is delegated to Auth0: no passwords are stored. Contact data of newcomers lives in its own
-- table so the API can keep it hidden until an agency unlocks it (audited).

CREATE TYPE user_role AS ENUM ('newcomer','hr','agency','admin');
CREATE TYPE property_type AS ENUM ('apartment','room','house','studio');
CREATE TYPE relocation_stage AS ENUM ('invited','profile_done','searching','visiting','contract_signed','moved_in','cancelled');

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  auth0_sub TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL,
  role user_role NOT NULL,
  locale TEXT NOT NULL DEFAULT 'en' CHECK (locale IN ('en','es','pt')),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE consents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose TEXT NOT NULL CHECK (purpose IN ('terms','privacy','share_with_agencies')),
  granted BOOLEAN NOT NULL,
  policy_version TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE companies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  industry TEXT NOT NULL DEFAULT '',
  city TEXT NOT NULL DEFAULT '',
  country CHAR(2) NOT NULL,
  allow_agency_contact BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE hr_profiles (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  full_name TEXT NOT NULL,
  phone TEXT NOT NULL DEFAULT ''
);

CREATE TABLE agencies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  city TEXT NOT NULL DEFAULT '',
  country CHAR(2) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  verified_at TIMESTAMPTZ,
  contact_credits INT NOT NULL DEFAULT 5 CHECK (contact_credits >= 0),
  credits_period DATE NOT NULL DEFAULT date_trunc('month', now())::date,
  plan TEXT NOT NULL DEFAULT 'free' CHECK (plan IN ('free','pro','enterprise')),
  plan_active_until TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE agency_profiles (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  full_name TEXT NOT NULL
);

CREATE TABLE newcomers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID UNIQUE REFERENCES users(id) ON DELETE SET NULL,   -- NULL until the newcomer claims the HR invitation
  company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  first_name TEXT NOT NULL,
  city_of_contract TEXT NOT NULL,
  city_key TEXT NOT NULL,                                        -- accent-folded lower-case city, for matching
  arrival_date DATE,
  budget_min NUMERIC(14,2) CHECK (budget_min >= 0),
  budget_max NUMERIC(14,2) CHECK (budget_max >= 0),
  currency CHAR(3) NOT NULL DEFAULT 'COP',
  bedrooms SMALLINT CHECK (bedrooms BETWEEN 0 AND 20),
  furnished BOOLEAN,
  preferred_districts TEXT[] NOT NULL DEFAULT '{}',
  stage relocation_stage NOT NULL DEFAULT 'invited',
  allow_contact BOOLEAN NOT NULL DEFAULT FALSE,                  -- opt-in, never opt-out
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX newcomers_open_idx ON newcomers (city_key, created_at DESC) WHERE allow_contact AND deleted_at IS NULL;

CREATE TABLE newcomer_contacts (
  newcomer_id UUID PRIMARY KEY REFERENCES newcomers(id) ON DELETE CASCADE,
  full_name TEXT NOT NULL,
  email TEXT NOT NULL,
  phone TEXT NOT NULL DEFAULT ''
);
CREATE INDEX newcomer_contacts_email_idx ON newcomer_contacts (lower(email));

CREATE TABLE properties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  price NUMERIC(14,2) NOT NULL CHECK (price >= 0),
  currency CHAR(3) NOT NULL DEFAULT 'COP',
  type property_type NOT NULL,
  bedrooms SMALLINT CHECK (bedrooms BETWEEN 0 AND 20),
  furnished BOOLEAN NOT NULL DEFAULT FALSE,
  city TEXT NOT NULL,
  city_key TEXT NOT NULL,
  district TEXT NOT NULL DEFAULT '',
  address TEXT NOT NULL DEFAULT '',
  available_from DATE,
  verified_at TIMESTAMPTZ,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX properties_city_idx ON properties (city_key, price) WHERE is_active;
CREATE INDEX properties_agency_idx ON properties (agency_id);

CREATE TABLE favorites (
  newcomer_id UUID NOT NULL REFERENCES newcomers(id) ON DELETE CASCADE,
  property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (newcomer_id, property_id)
);

CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  newcomer_id UUID NOT NULL REFERENCES newcomers(id) ON DELETE CASCADE,
  match_score SMALLINT NOT NULL CHECK (match_score BETWEEN 0 AND 100),
  message TEXT NOT NULL DEFAULT '',
  read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (agency_id, newcomer_id)
);
CREATE INDEX notifications_agency_idx ON notifications (agency_id, created_at DESC);

CREATE TABLE contact_unlocks (
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  newcomer_id UUID NOT NULL REFERENCES newcomers(id) ON DELETE CASCADE,
  unlocked_by UUID NOT NULL REFERENCES users(id),
  used_credit BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (agency_id, newcomer_id)
);

CREATE TABLE partner_inquiries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name TEXT NOT NULL,
  email TEXT NOT NULL,
  phone TEXT NOT NULL DEFAULT '',
  country TEXT NOT NULL DEFAULT '',
  city TEXT NOT NULL DEFAULT '',
  collaboration_type TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE landing_leads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL, email TEXT NOT NULL, phone TEXT NOT NULL DEFAULT '',
  city TEXT NOT NULL DEFAULT '', country TEXT NOT NULL DEFAULT '', interest TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  action TEXT NOT NULL,
  table_name TEXT NOT NULL DEFAULT '',
  record_id TEXT NOT NULL DEFAULT '',
  changes JSONB NOT NULL DEFAULT '{}',
  ip TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX audit_logs_user_idx ON audit_logs (user_id, created_at);
