-- 📊 DIAGNOSTIC: Vérifier l'état des tables d'engagement

-- 1️⃣ Posts par type
SELECT 
  author_type,
  COUNT(*) as total_posts
FROM posts
GROUP BY author_type;

-- Résultat attendu:
-- universite | X
-- centre_formation | Y

---

-- 2️⃣ Engagements universités
SELECT 
  COUNT(*) as total_engagements,
  COUNT(DISTINCT universite_id) as unique_universites,
  COUNT(DISTINCT type) as engagement_types,
  COUNT(DISTINCT user_id) as unique_users
FROM engagements_universites;

-- Résultat attendu:
-- total_engagements: 5+ ✅
-- unique_universites: 2
-- engagement_types: 2 (like, view)
-- unique_users: 1

---

-- 3️⃣ Engagements centres 
SELECT 
  COUNT(*) as total_engagements,
  COUNT(DISTINCT centre_id) as unique_centres,
  COUNT(DISTINCT type) as engagement_types,
  COUNT(DISTINCT user_id) as unique_users
FROM engagements_centres_formation;

-- Résultat attendu:
-- Si 0 posts de centre_formation → sera vide ✅ NORMAL
-- Si posts de centre_formation → devrait avoir données

---

-- 4️⃣ Détail engagements par entité
SELECT 
  universite_id,
  COUNT(*) as engagement_count,
  STRING_AGG(DISTINCT type, ', ') as types
FROM engagements_universites
GROUP BY universite_id
ORDER BY engagement_count DESC;

-- Résultat:
-- universite_id | engagement_count | types
-- 232cf2c9-... | 3 | view, like
-- 64b6d0f5-... | 2 | view, like

---

-- 5️⃣ Bonus: Impact sur PORA scoring
SELECT 
  u.id,
  u.nom,
  u.score_pora,
  (SELECT COUNT(*) FROM engagements_universites WHERE universite_id = u.id) as engagement_count,
  (SELECT COUNT(*) FROM followers_universites WHERE universite_id = u.id) as followers_count
FROM universites u
ORDER BY engagement_count DESC
LIMIT 10;

---

-- 6️⃣ Orientation recommendations
SELECT
  target_type,
  COUNT(*) as total_rows,
  COUNT(DISTINCT user_id) as unique_users,
  COUNT(DISTINCT target_id) as unique_targets,
  ROUND(AVG(score)::numeric, 4) as avg_score
FROM orientation_recommendations
GROUP BY target_type
ORDER BY target_type;

-- Résultat attendu après un quiz:
-- universite | 1+ ligne
-- centre     | 0+ ligne (selon le flux appelé)

---

-- 7️⃣ Dernières recommandations persistées
SELECT
  user_id,
  target_type,
  target_id,
  score,
  created_at
FROM orientation_recommendations
ORDER BY created_at DESC
LIMIT 20;

---

-- 8️⃣ Vue agrégée utilisée par PORA
SELECT *
FROM orientation_scores_universites
ORDER BY score_total DESC
LIMIT 10;

SELECT *
FROM orientation_scores_centres
ORDER BY score_total DESC
LIMIT 10;
