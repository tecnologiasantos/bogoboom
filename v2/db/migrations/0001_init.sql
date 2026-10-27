-- Bogoboom v1 schema (PostgreSQL 15+). Improves db/schema.v0-from-doc.sql:
--  * no password_hash (auth delegated to Cognito/OIDC; store external subject id)
--  * enums, NOT NULL/UNIQUE/FK actions, indexes, updated_at, soft delete
--  * contact data kept in separate tables so API can enforce "blurred until subscribed" SERVER-side
--  * Favorites uniqueness, currency, geo columns, consent tracking
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TYPE user_role AS ENUM ('newcomer','hr','agency','admin');
CREATE TYPE property_type AS ENUM ('apartment','room','house','studio');
CREATE TYPE sub_status AS ENUM ('trialing','active','past_due','cancelled','expired');
CREATE TYPE notif_type AS ENUM ('new_search','update','interest_shown','match');
CREATE TYPE relocation_stage AS ENUM ('invited','profile_done','searching','visiting','contract_signed','moved_in','cancelled');

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  auth_sub TEXT UNIQUE NOT NULL,              -- Cognito "sub"
  email CITEXT UNIQUE NOT NULL,
  role user_role NOT NULL,
  locale TEXT NOT NULL DEFAULT 'en' CHECK (locale IN ('en','es','pt')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE consents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose TEXT NOT NULL,                      -- 'terms','privacy','share_with_agencies','marketing'
  granted BOOLEAN NOT NULL,
  policy_version TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE companies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL, industry TEXT, city TEXT, country CHAR(2),
  allow_agency_contact BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE hr_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  full_name TEXT NOT NULL, phone_number TEXT
);

CREATE TABLE agencies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL, city TEXT, country CHAR(2), description TEXT,
  verified_at TIMESTAMPTZ,                    -- "Verified agency" badge
  contact_credits INT NOT NULL DEFAULT 5,     -- freemium: 5 free contacts / month
  credits_period DATE NOT NULL DEFAULT date_trunc('month', now())::date,  -- month the credits belong to (lazy monthly reset)
  plan TEXT NOT NULL DEFAULT 'free' CHECK (plan IN ('free','pro','enterprise')),
  plan_active_until TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE agency_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  full_name TEXT NOT NULL
);

-- Public/blurred part of a newcomer (what unsubscribed agencies may see)
CREATE TABLE newcomers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID UNIQUE REFERENCES users(id) ON DELETE SET NULL,   -- NULL until the newcomer claims the invite
  company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  first_name TEXT NOT NULL,
  city_of_contract TEXT NOT NULL,
  arrival_date DATE,
  budget_min NUMERIC(12,2), budget_max NUMERIC(12,2), currency CHAR(3) NOT NULL DEFAULT 'COP',
  bedrooms SMALLINT, furnished BOOLEAN, preferred_districts TEXT[] NOT NULL DEFAULT '{}',
  preferences JSONB NOT NULL DEFAULT '{}',
  stage relocation_stage NOT NULL DEFAULT 'invited',
  allow_contact BOOLEAN NOT NULL DEFAULT FALSE,   -- opt-in, not opt-out
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);
-- Private part: only returned by API when agency has active subscription/credit AND allow_contact
CREATE TABLE newcomer_contacts (
  newcomer_id UUID PRIMARY KEY REFERENCES newcomers(id) ON DELETE CASCADE,
  full_name TEXT, email TEXT, phone TEXT
);

CREATE TABLE properties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  title TEXT NOT NULL, description TEXT,
  price NUMERIC(12,2) NOT NULL CHECK (price >= 0), currency CHAR(3) NOT NULL DEFAULT 'COP',
  type property_type NOT NULL, bedrooms SMALLINT, furnished BOOLEAN NOT NULL DEFAULT FALSE,
  city TEXT NOT NULL, district TEXT, address TEXT, lat DOUBLE PRECISION, lng DOUBLE PRECISION,
  available_from DATE, verified_at TIMESTAMPTZ, is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE property_media (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
  s3_key TEXT NOT NULL, kind TEXT NOT NULL CHECK (kind IN ('image','contract','floorplan')), position INT NOT NULL DEFAULT 0
);

CREATE TABLE favorites (
  newcomer_id UUID NOT NULL REFERENCES newcomers(id) ON DELETE CASCADE,
  property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (newcomer_id, property_id)
);

CREATE TABLE subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  plan TEXT NOT NULL CHECK (plan IN ('free','pro','enterprise')),
  status sub_status NOT NULL,
  provider TEXT NOT NULL, provider_subscription_id TEXT UNIQUE,
  current_period_end TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  newcomer_id UUID REFERENCES newcomers(id) ON DELETE CASCADE,
  type notif_type NOT NULL DEFAULT 'new_search',
  match_score SMALLINT CHECK (match_score BETWEEN 0 AND 100),
  message TEXT, read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (agency_id, newcomer_id, type)       -- dedupe fan-out
);

CREATE TABLE contact_unlocks (               -- every reveal of private data is logged (credits + audit)
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id), newcomer_id UUID NOT NULL REFERENCES newcomers(id),
  unlocked_by UUID NOT NULL REFERENCES users(id), used_credit BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE (agency_id, newcomer_id)
);

CREATE TABLE conversations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id), newcomer_id UUID NOT NULL REFERENCES newcomers(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE (agency_id, newcomer_id)
);
CREATE TABLE messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  sender_id UUID NOT NULL REFERENCES users(id), content TEXT NOT NULL,
  read_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
  author_id UUID NOT NULL REFERENCES users(id), rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
  body TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE (agency_id, author_id)
);

CREATE TABLE uploads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  s3_key TEXT NOT NULL, file_type TEXT, description TEXT, scanned_at TIMESTAMPTZ,
  uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY, user_id UUID REFERENCES users(id), action TEXT NOT NULL,
  table_name TEXT, record_id UUID, changes JSONB, ip INET, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON properties (city, type, price) WHERE is_active;
CREATE INDEX ON properties (agency_id);
CREATE INDEX ON newcomers (city_of_contract, arrival_date) WHERE deleted_at IS NULL;
CREATE INDEX ON notifications (agency_id, read_at, created_at DESC);
CREATE INDEX ON messages (conversation_id, created_at);
CREATE INDEX ON audit_logs (user_id, created_at);
