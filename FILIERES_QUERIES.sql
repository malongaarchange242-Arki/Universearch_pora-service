-- =========================================================
-- SQL QUERIES FOR FETCHING LINKED FILIERES
-- =========================================================

-- This file explains the SQL queries used to retrieve 
-- filieres linked to universites and centres

-- =========================================================
-- 1. FETCH FILIERES FOR SPECIFIC UNIVERSITES
-- =========================================================

/*
Query to get all filieres linked to one or more universities.
Used in: PostUniversiteRecommendations handler

Input: List of universite_ids
Output: List of filiere names (unique, no duplicates)
*/

SELECT DISTINCT f.nom
FROM universite_filieres uf
JOIN filieres f ON f.id = uf.filiere_id
WHERE uf.universite_id IN (
    -- Replace with actual universite IDs
    'uuid-1',
    'uuid-2',
    'uuid-3'
)
ORDER BY f.nom;

-- Example with Supabase REST API filter:
-- GET /rest/v1/universite_filieres?select=filieres(nom)&universite_id=in.(uuid1,uuid2,uuid3)

-- =========================================================
-- 2. FETCH FILIERES FOR SPECIFIC CENTRES
-- =========================================================

/*
Query to get all filieres linked to one or more centres.
Used in: PostCentreRecommendations handler

Input: List of centre_ids
Output: List of filiere names (unique, no duplicates)
*/

SELECT DISTINCT f.nom
FROM centres_formation_filieres cff
JOIN filieres f ON f.id = cff.filiere_id
WHERE cff.centre_id IN (
    -- Replace with actual centre IDs
    'uuid-1',
    'uuid-2',
    'uuid-3'
)
ORDER BY f.nom;

-- Example with Supabase REST API filter:
-- GET /rest/v1/centres_formation_filieres?select=filieres(nom)&centre_id=in.(uuid1,uuid2,uuid3)

-- =========================================================
-- 3. COMPREHENSIVE QUERY - GET UNIVERSITE + FILIERES
-- =========================================================

/*
Get universities WITH their linked filieres in one query
(For reference/optimization analysis)
*/

SELECT 
    u.id,
    u.nom as universite_nom,
    json_agg(DISTINCT f.nom) as filieres
FROM universites u
LEFT JOIN universite_filieres uf ON u.id = uf.universite_id
LEFT JOIN filieres f ON f.id = uf.filiere_id
WHERE u.id IN ('uuid-1', 'uuid-2')
GROUP BY u.id, u.nom;

-- =========================================================
-- 4. COMPREHENSIVE QUERY - GET CENTRE + FILIERES
-- =========================================================

/*
Get centres WITH their linked filieres in one query
(For reference/optimization analysis)
*/

SELECT 
    c.id,
    c.nom as centre_nom,
    json_agg(DISTINCT f.nom) as filieres
FROM centres_formation c
LEFT JOIN centres_formation_filieres cff ON c.id = cff.centre_id
LEFT JOIN filieres f ON f.id = cff.filiere_id
WHERE c.id IN ('uuid-1', 'uuid-2')
GROUP BY c.id, c.nom;

-- =========================================================
-- 5. TABLE STRUCTURES
-- =========================================================

-- universite_filieres table
/*
CREATE TABLE universite_filieres (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    universite_id uuid NOT NULL REFERENCES universites(id) ON DELETE CASCADE,
    filiere_id uuid NOT NULL REFERENCES filieres(id) ON DELETE CASCADE,
    date_ajout timestamp DEFAULT now(),
    UNIQUE(universite_id, filiere_id)
);

CREATE INDEX idx_universite_filieres_universite ON universite_filieres (universite_id);
CREATE INDEX idx_universite_filieres_filiere ON universite_filieres (filiere_id);
*/

-- centres_formation_filieres table
/*
CREATE TABLE centres_formation_filieres (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    centre_id uuid NOT NULL REFERENCES centres_formation(id) ON DELETE CASCADE,
    filiere_id uuid NOT NULL REFERENCES filieres(id) ON DELETE CASCADE,
    date_ajout timestamp DEFAULT now(),
    UNIQUE(centre_id, filiere_id)
);

CREATE INDEX idx_centres_formation_filieres_centre ON centres_formation_filieres (centre_id);
CREATE INDEX idx_centres_formation_filieres_filiere ON centres_formation_filieres (filiere_id);
*/

-- =========================================================
-- 6. GO IMPLEMENTATION - fetchFilieresForUniversites
-- =========================================================

/*
Function that implements Query #1 in Go/Supabase REST API:

func fetchFilieresForUniversites(universiteIDs []string) ([]string, error) {
    // Build IN filter: in.(uuid1,uuid2,uuid3)
    
    u, _ := url.Parse(SupabaseURL + "/rest/v1/universite_filieres")
    q := u.Query()
    q.Set("select", "filieres(nom)")
    q.Set("universite_id", fmt.Sprintf("in.(%s)", inFilter.String()))
    q.Set("limit", "1000")
    u.RawQuery = q.Encode()
    
    // Parse response and extract unique noms
    // Return []string with filiere names
}

REST API Request Example:
GET /rest/v1/universite_filieres?select=filieres(nom)&universite_id=in.("uuid1","uuid2","uuid3")&limit=1000

Response Example:
[
  { "filieres": { "nom": "Génie Informatique" } },
  { "filieres": { "nom": "Data Science" } },
  { "filieres": { "nom": "Génie Informatique" } }  <- duplicate
]

After deduplication:
["Génie Informatique", "Data Science"]
*/

-- =========================================================
-- 7. GO IMPLEMENTATION - fetchFilieresForCentres
-- =========================================================

/*
Function that implements Query #2 in Go/Supabase REST API:

func fetchFilieresForCentres(centreIDs []string) ([]string, error) {
    // Build IN filter: in.(uuid1,uuid2,uuid3)
    
    u, _ := url.Parse(SupabaseURL + "/rest/v1/centres_formation_filieres")
    q := u.Query()
    q.Set("select", "filieres(nom)")
    q.Set("centre_id", fmt.Sprintf("in.(%s)", inFilter.String()))
    q.Set("limit", "1000")
    u.RawQuery = q.Encode()
    
    // Parse response and extract unique noms
    // Return []string with filiere names
}

REST API Request Example:
GET /rest/v1/centres_formation_filieres?select=filieres(nom)&centre_id=in.("uuid1","uuid2","uuid3")&limit=1000

Response Example:
[
  { "filieres": { "nom": "Développement Web" } },
  { "filieres": { "nom": "Marketing Digital" } }
]

After deduplication:
["Développement Web", "Marketing Digital"]
*/

-- =========================================================
-- 8. RESPONSE FORMAT
-- =========================================================

/*
Handler Response (PostUniversiteRecommendations):

{
  "universites": [
    {
      "id": "uuid1",
      "nom": "Université de Kinshasa",
      "score_pora": 0.85,
      ...
    }
  ],
  "univFilieres": [
    "Génie Informatique",
    "Data Science",
    "Développement Web"
  ]
}

Handler Response (PostCentreRecommendations):

{
  "centres": [
    {
      "id": "uuid1",
      "nom": "Centre de Formation XYZ",
      "score_pora": 0.72,
      ...
    }
  ],
  "centreFilieres": [
    "Développement Web",
    "Marketing Digital"
  ]
}
*/

-- =========================================================
-- 9. OPTIMIZATION NOTES
-- =========================================================

/*
✅ Performance Considerations:
- Both tables have indexes on their ID columns
- Joins are efficiently done via indexed foreign keys
- DISTINCT ensures no duplicates in response
- Limit 1000 prevents excessive data transfer

✅ Error Handling:
- If no filieres found, returns empty []string
- If query fails (HTTP error), logs warning and returns empty []string
- Never crashes the handler

✅ Deduplication:
- Uses Go map[string]bool to avoid duplicates
- Preserves only filiere names (nom field)
- Converts map to slice for JSON response
*/
