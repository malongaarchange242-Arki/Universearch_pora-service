-- =========================================================
-- CRÉER LES TABLES DE RELATIONS FILIÈRES
-- =========================================================

-- Créer la table universite_filieres si elle n'existe pas
create table if not exists universite_filieres (
  id uuid primary key default gen_random_uuid(),
  universite_id uuid not null references universites(id) on delete cascade,
  filiere_id uuid not null references filieres(id) on delete cascade,
  date_ajout timestamp default now(),
  unique(universite_id, filiere_id)
);

create index if not exists idx_universite_filieres_universite on universite_filieres (universite_id);
create index if not exists idx_universite_filieres_filiere on universite_filieres (filiere_id);

-- Créer la table centres_formation_filieres si elle n'existe pas
create table if not exists centres_formation_filieres (
  id uuid primary key default gen_random_uuid(),
  centre_id uuid not null references centres_formation(id) on delete cascade,
  filiere_id uuid not null references filieres(id) on delete cascade,
  date_ajout timestamp default now(),
  unique(centre_id, filiere_id)
);

create index if not exists idx_centres_formation_filieres_centre on centres_formation_filieres (centre_id);
create index if not exists idx_centres_formation_filieres_filiere on centres_formation_filieres (filiere_id);

-- =========================================================
-- PEUPLER LES RELATIONS AVEC DES DONNÉES DE TEST
-- =========================================================

-- RELATION: Université de Kinshasa → Filières Informatique + Sciences
insert into universite_filieres (universite_id, filiere_id)
select 
  u.id,
  f.id
from universites u
cross join filieres f
where u.nom = 'Université de Kinshasa'
  and (f.nom ilike '%informatique%' or f.nom ilike '%génie informatique%' or f.nom ilike '%développement%')
on conflict do nothing;

-- RELATION: Université Protestante → Filières Économie + Gestion
insert into universite_filieres (universite_id, filiere_id)
select 
  u.id,
  f.id
from universites u
cross join filieres f
where u.nom = 'Université Protestante au Congo'
  and (f.nom ilike '%gestion%' or f.nom ilike '%management%' or f.nom ilike '%commerce%' or f.nom ilike '%économie%')
on conflict do nothing;

-- RELATION: Université Catholique → Filières Médecine + Sciences
insert into universite_filieres (universite_id, filiere_id)
select 
  u.id,
  f.id
from universites u
cross join filieres f
where u.nom = 'Université Catholique du Congo'
  and (f.nom ilike '%médecine%' or f.nom ilike '%santé%' or f.nom ilike '%sciences%' or f.nom ilike '%biologie%')
on conflict do nothing;

-- RELATION: Université Kongo → Filières Droit + Commerce
insert into universite_filieres (universite_id, filiere_id)
select 
  u.id,
  f.id
from universites u
cross join filieres f
where u.nom = 'Université Kongo'
  and (f.nom ilike '%droit%' or f.nom ilike '%justice%' or f.nom ilike '%commerce%' or f.nom ilike '%communication%')
on conflict do nothing;

-- =========================================================
-- CENTRES DE FORMATION - RELATIONS FILIÈRES
-- =========================================================

-- RELATION: CFPT Kinshasa → Filières Informatique
insert into centres_formation_filieres (centre_id, filiere_id)
select 
  c.id,
  f.id
from centres_formation c
cross join filieres f
where c.nom = 'CFPT Kinshasa'
  and (f.nom ilike '%informatique%' or f.nom ilike '%développement%' or f.nom ilike '%réseaux%')
on conflict do nothing;

-- RELATION: Institut Supérieur de Technologie → Filières Technologie + Électronique
insert into centres_formation_filieres (centre_id, filiere_id)
select 
  c.id,
  f.id
from centres_formation c
cross join filieres f
where c.nom = 'Institut Supérieur de Technologie'
  and (f.nom ilike '%électronique%' or f.nom ilike '%technologie%' or f.nom ilike '%système%' or f.nom ilike '%réseaux%')
on conflict do nothing;

-- RELATION: Centre Digital Congo → Filières développement web + Data
insert into centres_formation_filieres (centre_id, filiere_id)
select 
  c.id,
  f.id
from centres_formation c
cross join filieres f
where c.nom = 'Centre Digital Congo'
  and (f.nom ilike '%développement%' or f.nom ilike '%web%' or f.nom ilike '%digital%' or f.nom ilike '%data%')
on conflict do nothing;

-- =========================================================
-- VÉRIFICATION
-- =========================================================
select 'Relations Universités-Filières' as table_name, count(*) as total
from universite_filieres
union all
select 'Relations Centres-Filières' as table_name, count(*) as total
from centres_formation_filieres;
