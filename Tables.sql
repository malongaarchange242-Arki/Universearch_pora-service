-- =========================================================
-- EXTENSIONS
-- =========================================================
create extension if not exists "pgcrypto";

-- =========================================================
-- UNIVERSITES
-- =========================================================
create table universites (
  id uuid primary key default gen_random_uuid(),
  nom text not null,
  description text,
  contacts text,
  email text,
  lien_site text,
  logo_url text,
  couverture_logo_url text,
  domaine text,
  bde_id uuid,
  statut text default 'inactive',
  score_pora float8 default 0,
  video_url text,
  date_creation timestamp default now()
);

-- =========================================================
-- BDE
-- =========================================================
create table bde (
  id uuid primary key default gen_random_uuid(),
  universite_id uuid references universites(id) on delete cascade,
  nom text not null,
  description text,
  logo_url text,
  video_url text,
  date_creation timestamp default now()
);

-- =========================================================
-- FOLLOWERS UNIVERSITES
-- =========================================================
create table followers_universites (
  user_id uuid not null,
  universite_id uuid not null references universites(id) on delete cascade,
  date_follow timestamp default now(),
  primary key (user_id, universite_id)
);

create index idx_followers_universites_universite
on followers_universites (universite_id);

-- =========================================================
-- ENGAGEMENTS UNIVERSITES
-- =========================================================
create table engagements_universites (
  id uuid primary key default gen_random_uuid(),
  universite_id uuid references universites(id) on delete cascade,
  type text not null,
  user_id uuid,
  post_id uuid,
  date timestamp default now()
);

create index idx_engagements_universites_universite
on engagements_universites (universite_id);

-- =========================================================
-- RECOMMANDATIONS UNIVERSITES (LEGACY / HORS PORA)
-- =========================================================
create table universites_recommandations (
  from_universite_id uuid references universites(id) on delete cascade,
  to_universite_id uuid references universites(id) on delete cascade,
  poids float8 default 0,
  raison text,
  primary key (from_universite_id, to_universite_id)
);

-- =========================================================
-- CENTRES DE FORMATION PROFESSIONNELLE
-- =========================================================
create table centres_formation (
  id uuid primary key default gen_random_uuid(),
  nom text not null,
  description text,
  contacts text,
  email text,
  lien_site text,
  logo_url text,
  couverture_logo_url text,
  domaine text,
  statut text default 'inactive',
  score_pora float8 default 0,
  video_url text,
  date_creation timestamp default now()
);

-- =========================================================
-- FOLLOWERS CENTRES DE FORMATION
-- =========================================================
create table followers_centres_formation (
  user_id uuid not null,
  centre_id uuid not null references centres_formation(id) on delete cascade,
  date_follow timestamp default now(),
  primary key (user_id, centre_id)
);

create index idx_followers_centres_centre
on followers_centres_formation (centre_id);

-- =========================================================
-- ENGAGEMENTS CENTRES DE FORMATION
-- =========================================================
create table engagements_centres_formation (
  id uuid primary key default gen_random_uuid(),
  centre_id uuid references centres_formation(id) on delete cascade,
  type text not null,
  user_id uuid,
  post_id uuid,
  date timestamp default now()
);

-- =========================================================
-- DOMAINES & FILIÈRES POUR CENTRES
-- =========================================================

create table domaines_centre (
    id uuid primary key default gen_random_uuid(),
    nom text not null unique,
    description text,
    created_at timestamp default now()
);

create table filieres_centre (
    id uuid primary key default gen_random_uuid(),
    nom text not null,
    description text,
    domaine_id uuid references domaines_centre(id) on delete cascade,
    created_at timestamp default now()
);

create table centre_formation_filieres (
    id uuid primary key default gen_random_uuid(),
    centre_id uuid references centres_formation(id) on delete cascade,
    filiere_id uuid references filieres_centre(id) on delete cascade,
    created_at timestamp default now(),
    constraint uniq_centre_filiere unique (centre_id, filiere_id)
);

create index idx_engagements_centres_centre
on engagements_centres_formation (centre_id);

-- =========================================================
-- RECOMMANDATIONS CENTRES (LEGACY / HORS PORA)
-- =========================================================
create table centres_formation_recommandations (
  from_centre_id uuid references centres_formation(id) on delete cascade,
  to_centre_id uuid references centres_formation(id) on delete cascade,
  poids float8 default 0,
  raison text,
  primary key (from_centre_id, to_centre_id)
);

-- =========================================================
-- RECOMMANDATIONS CROISÉES
-- Université ↔ Centre de formation
-- SIGNAL TRANSVERSAL PORA
-- =========================================================
create table formation_recommandations_cross (
  from_type text check (from_type in ('universite', 'centre')),
  from_id uuid not null,
  to_type text check (to_type in ('universite', 'centre')),
  to_id uuid not null,
  poids float8 default 0,
  raison text,
  primary key (from_type, from_id, to_type, to_id)
);

-- 🔒 Sécurité des poids (anti-biais)
alter table formation_recommandations_cross
add constraint chk_cross_poids_range
check (poids >= 0 and poids <= 100);

-- 🔎 Index performance
create index idx_cross_to
on formation_recommandations_cross (to_type, to_id);

create index idx_cross_from
on formation_recommandations_cross (from_type, from_id);


-- =========================================================
-- ORIENTATION : RÉPONSES BRUTES AU QUIZ
-- =========================================================
create table orientation_quiz_responses (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  quiz_version text not null,
  responses jsonb not null,
  created_at timestamp default now()
);

create index idx_orientation_quiz_user
on orientation_quiz_responses (user_id);

-- =========================================================
-- ORIENTATION : FEATURES ML
-- =========================================================
create table orientation_features (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  quiz_version text not null,
  features jsonb not null,
  created_at timestamp default now()
);

create index idx_orientation_features_user
on orientation_features (user_id);

-- =========================================================
-- ORIENTATION : PROFIL FINAL
-- =========================================================
create table orientation_profiles (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  profile jsonb not null,
  confidence float8 default 0,
  engine text check (engine in ('rule', 'ml')),
  created_at timestamp default now()
);

create index idx_orientation_profile_user
on orientation_profiles (user_id);

-- =========================================================
-- ORIENTATION : FEEDBACK UTILISATEUR
-- =========================================================
create table orientation_feedback (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  satisfaction int check (satisfaction between 1 and 5),
  changed_orientation boolean default false,
  success boolean,
  created_at timestamp default now()
);


-- =========================================================
-- ORIENTATION : QUIZZ
-- =========================================================
create table orientation_quizzes (
  id uuid primary key default gen_random_uuid(),
  quiz_code text unique not null,      -- ex: "orientation_etudiant_v1"
  target_profile text check (
    target_profile in ('eleve', 'etudiant', 'parent')
  ) not null,
  version int not null,
  title text not null,
  description text,
  is_active boolean default true,
  created_at timestamp default now()
);


-- =========================================================
-- ORIENTATION : QUESTIONS
-- =========================================================
create table orientation_quiz_questions (
  id uuid primary key default gen_random_uuid(),
  quiz_id uuid references orientation_quizzes(id) on delete cascade,
  question_code text not null,          -- ex: "logic_1"
  question_text text not null,
  question_type text check (
    question_type in ('likert', 'choice', 'boolean')
  ) not null,
  order_index int not null,
  is_required boolean default true
);


-- =========================================================
-- ORIENTATION : QUESTION → FEATURES
-- =========================================================
create table orientation_question_feature_weights (
  question_id uuid references orientation_quiz_questions(id) on delete cascade,
  feature_name text not null,           -- ex: "logic_score"
  weight float8 not null,                -- ex: 0.25
  primary key (question_id, feature_name)
);

-- Sécurité des poids
alter table orientation_question_feature_weights
add constraint chk_feature_weight_range
check (weight >= 0 and weight <= 1);


ALTER TABLE orientation_quiz_questions
ADD COLUMN created_at timestamp DEFAULT now();

ALTER TABLE orientation_features
ADD COLUMN quiz_id uuid references orientation_quizzes(id);

ALTER TABLE orientation_profiles
ADD COLUMN quiz_version text;

ALTER TABLE orientation_profiles
ADD COLUMN explanation jsonb;


create table orientation_recommendations (
  id uuid primary key default gen_random_uuid(),

  user_id uuid not null,
  profile_id uuid references orientation_profiles(id) on delete cascade,

  target_type text check (target_type in ('universite', 'centre')),
  target_id uuid not null,

  score float8 not null,          -- score PROA (0 → 1)
  rank int not null,              -- rang dans les recommandations
  reason text,                    -- explication lisible
  confidence float8 default 0,    -- confiance locale

  created_at timestamp default now()
);

create index idx_orientation_reco_user
on orientation_recommendations (user_id);

create index idx_orientation_reco_target
on orientation_recommendations (target_type, target_id);


create table orientation_scores (
  id uuid primary key default gen_random_uuid(),

  user_id uuid not null,

  target_type text check (target_type in ('universite', 'centre')) not null,
  target_id uuid not null,

  score float8 not null check (score >= 0 and score <= 1),

  profile_id uuid references orientation_profiles(id) on delete cascade,
  quiz_id uuid references orientation_quizzes(id),

  created_at timestamp default now(),

  unique (user_id, target_type, target_id, quiz_id)
);

create index idx_orientation_scores_target
on orientation_scores (target_type, target_id);

create index idx_orientation_scores_user
on orientation_scores (user_id);








-- =============================================================
-- 🔹 SUPPRESSION DES ENREGISTREMENTS EXISTANTS
-- =============================================================
DELETE FROM followers_universites;
DELETE FROM engagements_universites;
DELETE FROM orientation_recommendations;
DELETE FROM formation_recommandations_cross;
DELETE FROM followers_centres_formation;
DELETE FROM engagements_centres_formation;
DELETE FROM orientation_profiles;
DELETE FROM universites;
DELETE FROM centres_formation;

-- =============================================================
-- 🔹 UNIVERSITÉS
-- =============================================================
INSERT INTO universites (id, nom, description, domaine, statut, score_pora)
VALUES
(gen_random_uuid(),'Université de Kinshasa','Grande université publique','Sciences','active',0),
(gen_random_uuid(),'Université Protestante au Congo','Université privée reconnue','Économie','active',0),
(gen_random_uuid(),'Université Catholique du Congo','Université catholique','Médecine','active',0),
(gen_random_uuid(),'Université Kongo','Université régionale','Droit','active',0);

-- =============================================================
-- 🔹 CENTRES DE FORMATION
-- =============================================================
INSERT INTO centres_formation (id, nom, description, domaine, statut, score_pora)
VALUES
(gen_random_uuid(),'Centre Alpha','Centre de formation informatique','Informatique','active',0),
(gen_random_uuid(),'Centre Beta','Centre de formation commerce','Commerce','active',0),
(gen_random_uuid(),'Centre Gamma','Centre de formation langues','Langues','active',0);

-- =============================================================
-- 🔹 FOLLOWERS UNIVERSITÉS
-- =============================================================
INSERT INTO followers_universites (user_id, universite_id, date_follow)
SELECT gen_random_uuid(), id, now() FROM universites;

-- =============================================================
-- 🔹 ENGAGEMENTS UNIVERSITÉS
-- =============================================================
INSERT INTO engagements_universites (id, universite_id, type, date)
SELECT gen_random_uuid(), id, 'like', now() FROM universites;

-- =============================================================
-- 🔹 ORIENTATION PROFILES
-- =============================================================
INSERT INTO orientation_profiles (id, user_id, profile, confidence, engine, created_at)
VALUES
(gen_random_uuid(), gen_random_uuid(), '{}', 0.9, 'rule', now()),
(gen_random_uuid(), gen_random_uuid(), '{}', 0.8, 'rule', now()),
(gen_random_uuid(), gen_random_uuid(), '{}', 1.0, 'rule', now()),
(gen_random_uuid(), gen_random_uuid(), '{}', 0.85, 'rule', now());

-- =============================================================
-- 🔹 ORIENTATION RECOMMENDATIONS
-- =============================================================
-- On récupère quelques universités pour associer les recommandations
WITH uni AS (
  SELECT id FROM universites ORDER BY nom
),
prof AS (
  SELECT id FROM orientation_profiles ORDER BY created_at
)
INSERT INTO orientation_recommendations (id, user_id, profile_id, target_type, target_id, score, confidence, rank, reason)
SELECT gen_random_uuid(),
       gen_random_uuid(),        -- user_id fictif
       prof.id,                 -- profile_id
       'universite',
       uni.id,
       0.8,
       0.9,
       ROW_NUMBER() OVER (),
       'Top match'
FROM uni, prof
LIMIT 4;

-- =============================================================
-- 🔹 RECOMMANDATIONS CROISÉES (Centre → Université)
-- =============================================================
WITH centres AS (
  SELECT id FROM centres_formation ORDER BY nom
),
universites AS (
  SELECT id FROM universites ORDER BY nom
)
INSERT INTO formation_recommandations_cross (from_type, from_id, to_type, to_id, poids, raison)
SELECT 'centre', centres.id, 'universite', universites.id, 0.7, 'Cross recommendation'
FROM centres, universites
LIMIT 4;

-- =============================================================
-- 🔹 FOLLOWERS CENTRES
-- =============================================================
INSERT INTO followers_centres_formation (user_id, centre_id, date_follow)
SELECT gen_random_uuid(), id, now() FROM centres_formation;

-- =============================================================
-- 🔹 ENGAGEMENTS CENTRES
-- =============================================================
INSERT INTO engagements_centres_formation (id, centre_id, type, date)
SELECT gen_random_uuid(), id, 'like', now() FROM centres_formation;


-- followers différents
INSERT INTO followers_universites (user_id, universite_id)
SELECT gen_random_uuid(), id FROM universites LIMIT 1;

INSERT INTO followers_universites (user_id, universite_id)
SELECT gen_random_uuid(), id FROM universites OFFSET 1 LIMIT 2;

-- engagements différents
INSERT INTO engagements_universites (universite_id, type)
SELECT id, 'like' FROM universites OFFSET 2 LIMIT 1;

create table orientation_events (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  target_type text check (target_type in ('universite', 'centre')),
  target_id uuid not null,
  event_type text check (
    event_type in ('view', 'click', 'save', 'apply', 'confirm')
  ),
  created_at timestamp default now()
);


create view orientation_scores_universites as
select
  target_id as universite_id,
  sum(
    case event_type
      when 'view' then 0.2
      when 'click' then 0.4
      when 'save' then 0.6
      when 'apply' then 0.8
      when 'confirm' then 1.0
    end
  ) as score
from orientation_events
where target_type = 'universite'
group by target_id;


create or replace view orientation_scores_centres as
select
  target_id as centre_id,
  sum(
    case event_type
      when 'view' then 0.2
      when 'click' then 0.4
      when 'save' then 0.6
      when 'apply' then 0.8
      when 'confirm' then 1.0
    end
  ) as score
from orientation_events
where target_type = 'centre'
group by target_id;


ALTER TABLE universites
ADD COLUMN score_pora_prev DOUBLE PRECISION DEFAULT 0;

ALTER TABLE centres_formation
ADD COLUMN score_pora_prev DOUBLE PRECISION DEFAULT 0;

UPDATE universites
SET score_pora_prev = score_pora;

UPDATE centres_formation
SET score_pora_prev = score_pora;




create table if not exists orientation_recommendations (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    target_id uuid not null,
    target_type text check (target_type in ('universite', 'centre')),
    score numeric(5,4) not null,
    engine text default 'pora',
    created_at timestamptz default now()
);


drop view if exists orientation_scores_universites;

create view orientation_scores_universites as
select
    target_id as universite_id,
    count(distinct user_id)          as nb_users,
    avg(score)::numeric(5,4)          as score_moyen,
    sum(score)::numeric(7,4)          as score_total
from orientation_recommendations
where target_type = 'universite'
group by target_id;


drop view if exists orientation_scores_centres;

create view orientation_scores_centres as
select
    target_id as centre_id,
    count(distinct user_id)          as nb_users,
    avg(score)::numeric(5,4)          as score_moyen,
    sum(score)::numeric(7,4)          as score_total
from orientation_recommendations
where target_type = 'centre'
group by target_id;


update centres_formation c
set score_pora = s.score_total
from orientation_scores_centres s
where c.id = s.centre_id;


update universites u
set score_pora = s.score_total
from orientation_scores_universites s
where u.id = s.universite_id;


ALTER TABLE orientation_recommendations
ALTER COLUMN rank DROP NOT NULL;

ALTER TABLE orientation_recommendations
ALTER COLUMN confidence DROP NOT NULL;

ALTER TABLE orientation_recommendations
ALTER COLUMN reason DROP NOT NULL;


alter table universites
add column score_details jsonb default '{}'::jsonb;

alter table centres_formation
add column score_details jsonb default '{}'::jsonb;
