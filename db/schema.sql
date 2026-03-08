-- services/pora-service/db/schema.sql

-- Table: universities
CREATE TABLE IF NOT EXISTS universities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  city text,
  country text,
  website text,
  description text,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_universities_slug ON universities(slug);
CREATE INDEX IF NOT EXISTS idx_universities_city ON universities(city);

-- Table: rankings
CREATE TABLE IF NOT EXISTS rankings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  university_id uuid NOT NULL,
  year integer NOT NULL,
  rank integer,
  source text,
  score numeric,
  created_at timestamptz DEFAULT now(),
  CONSTRAINT fk_rankings_university FOREIGN KEY(university_id)
    REFERENCES universities(id) ON DELETE CASCADE,
  CONSTRAINT uniq_univ_year_source UNIQUE (university_id, year, source)
);
CREATE INDEX IF NOT EXISTS idx_rankings_university ON rankings(university_id);
CREATE INDEX IF NOT EXISTS idx_rankings_year ON rankings(year);

-- Optional: departments / programs
CREATE TABLE IF NOT EXISTS departments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  university_id uuid NOT NULL,
  name text NOT NULL,
  code text,
  created_at timestamptz DEFAULT now(),
  CONSTRAINT fk_departments_university FOREIGN KEY(university_id)
    REFERENCES universities(id) ON DELETE CASCADE
);

-- Notes:
-- - Uses gen_random_uuid() (pgcrypto). Replace with uuid_generate_v4() if preferred.
-- - Adjust types/constraints to match your real data and FK targets.
