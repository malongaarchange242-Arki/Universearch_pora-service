
-- =========================================================
-- DONNÉES DE TEST PORA (UNIVERSITÉS & CENTRES)
-- =========================================================

-- UNIVERSITÉS
insert into universites (nom, description, domaine, statut) values
('Université de Kinshasa', 'Grande université publique', 'Sciences', 'active'),
('Université Protestante au Congo', 'Université privée reconnue', 'Économie', 'active'),
('Université Catholique du Congo', 'Université catholique', 'Médecine', 'active'),
('Université Kongo', 'Université régionale', 'Droit', 'active');

-- CENTRES DE FORMATION
insert into centres_formation (nom, description, domaine, statut) values
('CFPT Kinshasa', 'Centre de formation professionnelle', 'Informatique', 'active'),
('Institut Supérieur de Technologie', 'Formation technique', 'Électronique', 'active'),
('Centre Digital Congo', 'Formation numérique', 'Développement web', 'active');

-- FOLLOWERS UNIVERSITÉS (FAKE USERS)
insert into followers_universites (user_id, universite_id)
select gen_random_uuid(), id from universites;

-- ENGAGEMENTS UNIVERSITÉS
insert into engagements_universites (universite_id, type)
select id, 'like' from universites;

-- FOLLOWERS CENTRES
insert into followers_centres_formation (user_id, centre_id)
select gen_random_uuid(), id from centres_formation;

-- ENGAGEMENTS CENTRES
insert into engagements_centres_formation (centre_id, type)
select id, 'view' from centres_formation;
