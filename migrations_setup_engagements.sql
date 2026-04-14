-- =========================================================
-- 🔥 MIGRATION: Création des tables d'engagement PORA
-- =========================================================
-- Cette migration crée les tables engagements_universites 
-- et engagements_centres_formation que le content-service 
-- alimente automatiquement.
--
-- EXÉCUTER dans Supabase SQL Editor!
-- =========================================================

-- 1️⃣ Table: engagements_universites
CREATE TABLE IF NOT EXISTS engagements_universites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  universite_id UUID NOT NULL REFERENCES universites(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('like', 'comment', 'view', 'share')),
  user_id UUID,
  post_id UUID,
  date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les queries PORA (count par universite)
CREATE INDEX IF NOT EXISTS idx_engagements_universites_id 
  ON engagements_universites(universite_id);

CREATE INDEX IF NOT EXISTS idx_engagements_universites_type 
  ON engagements_universites(type);

CREATE INDEX IF NOT EXISTS idx_engagements_universites_date 
  ON engagements_universites(date DESC);

-- 2️⃣ Table: engagements_centres_formation
CREATE TABLE IF NOT EXISTS engagements_centres_formation (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  centre_id UUID NOT NULL REFERENCES centres_formation(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('like', 'comment', 'view', 'share')),
  user_id UUID,
  post_id UUID,
  date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les queries PORA (count par centre)
CREATE INDEX IF NOT EXISTS idx_engagements_centres_id 
  ON engagements_centres_formation(centre_id);

CREATE INDEX IF NOT EXISTS idx_engagements_centres_type 
  ON engagements_centres_formation(type);

CREATE INDEX IF NOT EXISTS idx_engagements_centres_date 
  ON engagements_centres_formation(date DESC);

-- 3️⃣ Enable RLS
ALTER TABLE engagements_universites ENABLE ROW LEVEL SECURITY;
ALTER TABLE engagements_centres_formation ENABLE ROW LEVEL SECURITY;

-- 4️⃣ RLS Policy: Anyone can READ (pour PORA)
CREATE POLICY read_engagements_universites ON engagements_universites
FOR SELECT USING (true);

CREATE POLICY read_engagements_centres ON engagements_centres_formation
FOR SELECT USING (true);

-- 5️⃣ RLS Policy: Content-service peut INSERT (via service role)
CREATE POLICY insert_engagements_universites ON engagements_universites
FOR INSERT WITH CHECK (true);

CREATE POLICY insert_engagements_centres ON engagements_centres_formation
FOR INSERT WITH CHECK (true);

-- =========================================================
-- ✅ Tables créées! PORA peut maintenant les interroger.
-- =========================================================
