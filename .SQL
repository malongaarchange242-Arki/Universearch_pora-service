CREATE TABLE users (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,

  nom VARCHAR(100) NOT NULL,
  prenom VARCHAR(100) NOT NULL,

  date_naissance DATE,
  genre VARCHAR(10),

  profil VARCHAR(20) NOT NULL CHECK (
    profil IN (
      'bachelier',
      'etudiant',
      'parent',
      'admin',
      'superviseur',
      'universite',
      'bde'
    )
  ),

  telephone VARCHAR(20),
  email VARCHAR(150) UNIQUE NOT NULL,

  role VARCHAR(30) NOT NULL,
  
  is_active BOOLEAN DEFAULT TRUE,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);


ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own profile"
ON users
FOR SELECT
USING (auth.uid() = id);


CREATE TABLE profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,

  nom VARCHAR(100) NOT NULL,
  prenom VARCHAR(100) NOT NULL,

  date_naissance DATE,
  genre VARCHAR(10),

  telephone VARCHAR(20),
  email VARCHAR(150) UNIQUE NOT NULL,

  profile_type VARCHAR(30) NOT NULL CHECK (
    profile_type IN (
      'utilisateur',
      'admin',
      'superviseur',
      'universite',
      'bde',
      'centre_formation'
    )
  ),

  is_active BOOLEAN DEFAULT TRUE,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE utilisateurs (
  id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,

  user_type VARCHAR(20) NOT NULL CHECK (
    user_type IN ('bachelier', 'etudiant', 'parent')
  ),

  niveau_etude VARCHAR(50),
  filiere_actuelle VARCHAR(100),

  created_at TIMESTAMP DEFAULT NOW()
);


CREATE TABLE admins (
  id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,

  can_manage_universities BOOLEAN DEFAULT TRUE,
  can_manage_payments BOOLEAN DEFAULT TRUE,

  created_at TIMESTAMP DEFAULT NOW()
);


CREATE TABLE superviseurs (
  id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,

  can_manage_admins BOOLEAN DEFAULT TRUE,
  can_view_transactions BOOLEAN DEFAULT TRUE,

  created_at TIMESTAMP DEFAULT NOW()
);
ALTER TABLE universites
ADD COLUMN profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE;

ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE utilisateurs ENABLE ROW LEVEL SECURITY;
ALTER TABLE admins ENABLE ROW LEVEL SECURITY;
ALTER TABLE superviseurs ENABLE ROW LEVEL SECURITY;

CREATE POLICY "User can read own profile"
ON profiles
FOR SELECT
USING (auth.uid() = id);

CREATE POLICY "Admin can view all profiles"
ON profiles
FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM admins
    WHERE admins.id = auth.uid()
  )
);


CREATE POLICY "Superviseur full access"
ON profiles
FOR ALL
USING (
  EXISTS (
    SELECT 1 FROM superviseurs
    WHERE superviseurs.id = auth.uid()
  )
);


CREATE OR REPLACE FUNCTION create_profile_account(
  p_user_id UUID,
  p_nom TEXT,
  p_prenom TEXT,
  p_email TEXT,
  p_profile_type TEXT,
  p_user_type TEXT DEFAULT NULL -- pour utilisateurs
)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  -- Sécurité : seuls certains rôles sont créables publiquement
  IF p_profile_type NOT IN ('utilisateur') THEN
    RAISE EXCEPTION 'Unauthorized profile type';
  END IF;

  -- Insert profile
  INSERT INTO profiles (
    id, nom, prenom, email, profile_type
  )
  VALUES (
    p_user_id, p_nom, p_prenom, p_email, p_profile_type
  );

  -- Insert utilisateur
  IF p_profile_type = 'utilisateur' THEN
    INSERT INTO utilisateurs (
      id, user_type
    )
    VALUES (
      p_user_id, p_user_type
    );
  END IF;

END;
$$;


CREATE OR REPLACE FUNCTION create_admin_account(
  p_user_id UUID,
  p_nom TEXT,
  p_prenom TEXT,
  p_email TEXT
)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  -- Seul un superviseur peut créer un admin
  IF NOT EXISTS (
    SELECT 1 FROM superviseurs WHERE id = auth.uid()
  ) THEN
    RAISE EXCEPTION 'Only superviseur can create admin';
  END IF;

  INSERT INTO profiles (id, nom, prenom, email, profile_type)
  VALUES (p_user_id, p_nom, p_prenom, p_email, 'admin');

  INSERT INTO admins (id)
  VALUES (p_user_id);
END;
$$;


-- ================================
-- Universearch Identity CLEAN SCRIPT
-- ================================

-- 1️⃣ Supprimer les POLICIES (si elles existent)
DROP POLICY IF EXISTS "User read own profile" ON profiles;
DROP POLICY IF EXISTS "Admin read all profiles" ON profiles;
DROP POLICY IF EXISTS "Superviseur full access profiles" ON profiles;

-- 2️⃣ Désactiver RLS (optionnel mais propre)
ALTER TABLE IF EXISTS superviseurs DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS admins DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS utilisateurs DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS profiles DISABLE ROW LEVEL SECURITY;

-- 3️⃣ Supprimer les TABLES (ordre inverse des dépendances)
DROP TABLE IF EXISTS superviseurs CASCADE;
DROP TABLE IF EXISTS admins CASCADE;
DROP TABLE IF EXISTS utilisateurs CASCADE;
DROP TABLE IF EXISTS profiles CASCADE;
