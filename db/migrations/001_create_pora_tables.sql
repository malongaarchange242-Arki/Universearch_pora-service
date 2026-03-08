-- services/pora-service/db/migrations/001_create_pora_tables.sql
-- Migration: create universities, rankings, departments
-- Generated: 2026-01-27

-- CREATE TABLE universities
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

-- CREATE TABLE rankings
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

-- CREATE TABLE departments
CREATE TABLE IF NOT EXISTS departments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  university_id uuid NOT NULL,
  name text NOT NULL,
  code text,
  created_at timestamptz DEFAULT now(),
  CONSTRAINT fk_departments_university FOREIGN KEY(university_id)
    REFERENCES universities(id) ON DELETE CASCADE
);
