# PORA - Filières Enhancement

## Overview

PORA (Recommendation Service) has been enhanced to return linked **filieres** (study programs) for recommended universities and centres de formation.

### Before
```json
{
  "universites": [...],
  "centres": [...]
}
```

### After
```json
{
  "universites": [...],
  "univFilieres": ["Génie Informatique", "Data Science"],
  "centres": [...],
  "centreFilieres": ["Développement Web", "Marketing Digital"]
}
```

---

## Changes Made

### 1. ✅ Added Functions to `supabase.go`

#### `fetchFilieresForUniversites(universiteIDs []string) ([]string, error)`
- Fetches all unique filieres linked to a list of universities
- Uses the `universite_filieres` join table
- Returns a deduplicated list of filiere names
- Handles errors gracefully (returns empty slice instead of failing)

**API Endpoint Used:**
```
GET /rest/v1/universite_filieres?select=filieres(nom)&universite_id=in.(uuid1,uuid2,uuid3)&limit=1000
```

#### `fetchFilieresForCentres(centreIDs []string) ([]string, error)`
- Fetches all unique filieres linked to a list of centres
- Uses the `centres_formation_filieres` join table
- Returns a deduplicated list of filiere names
- Same error handling as above

**API Endpoint Used:**
```
GET /rest/v1/centres_formation_filieres?select=filieres(nom)&centre_id=in.(uuid1,uuid2,uuid3)&limit=1000
```

---

### 2. ✅ Modified Handlers in `handlers.go`

#### `PostUniversiteRecommendations`
**Changes:**
1. After filtering universities, extract their IDs
2. Call `fetchFilieresForUniversites(universiteIDs)` 
3. Add `univFilieres` to JSON response
4. Add logging for debug: `📚 Filières universités: [...]`

**New Response:**
```json
{
  "universites": [...],
  "univFilieres": [...]
}
```

#### `PostCentreRecommendations`
**Changes:**
1. After filtering centres, extract their IDs
2. Call `fetchFilieresForCentres(centreIDs)`
3. Add `centreFilieres` to JSON response
4. Add logging for debug: `📚 Filières centres: [...]`

**New Response:**
```json
{
  "centres": [...],
  "centreFilieres": [...]
}
```

#### Empty Fields Case (Both handlers)
When no recommended fields are provided, both handlers now return:
```json
{
  "universites": [],
  "univFilieres": []
}
// or
{
  "centres": [],
  "centreFilieres": []
}
```

---

## Database Schema

### `universite_filieres` Table
```sql
CREATE TABLE universite_filieres (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    universite_id uuid NOT NULL REFERENCES universites(id) ON DELETE CASCADE,
    filiere_id uuid NOT NULL REFERENCES filieres(id) ON DELETE CASCADE,
    date_ajout timestamp DEFAULT now(),
    UNIQUE(universite_id, filiere_id)
);

CREATE INDEX idx_universite_filieres_universite ON universite_filieres (universite_id);
CREATE INDEX idx_universite_filieres_filiere ON universite_filieres (filiere_id);
```

### `centres_formation_filieres` Table
```sql
CREATE TABLE centres_formation_filieres (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    centre_id uuid NOT NULL REFERENCES centres_formation(id) ON DELETE CASCADE,
    filiere_id uuid NOT NULL REFERENCES filieres(id) ON DELETE CASCADE,
    date_ajout timestamp DEFAULT now(),
    UNIQUE(centre_id, filiere_id)
);

CREATE INDEX idx_centres_formation_filieres_centre ON centres_formation_filieres (centre_id);
CREATE INDEX idx_centres_formation_filieres_filiere ON centres_formation_filieres (filiere_id);
```

---

## SQL Queries

See `FILIERES_QUERIES.sql` for detailed query examples.

### Basic Query - Universities
```sql
SELECT DISTINCT f.nom
FROM universite_filieres uf
JOIN filieres f ON f.id = uf.filiere_id
WHERE uf.universite_id IN ('uuid-1', 'uuid-2', 'uuid-3')
ORDER BY f.nom;
```

### Basic Query - Centres
```sql
SELECT DISTINCT f.nom
FROM centres_formation_filieres cff
JOIN filieres f ON f.id = cff.filiere_id
WHERE cff.centre_id IN ('uuid-1', 'uuid-2', 'uuid-3')
ORDER BY f.nom;
```

---

## Request/Response Flow

### 1. Frontend sends quiz results to PROA
```
POST /orientation/quiz
{
  "responses": { "Q1": 5, "Q2": 3, ... }
}
```

### 2. PROA returns recommended fields
```json
{
  "recommended_fields": [
    { "field_name": "Génie Informatique", "score": 0.92 },
    { "field_name": "Data Science", "score": 0.85 }
  ]
}
```

### 3. Frontend sends fields to PORA
```
POST /recommendations/universites
{
  "user_id": "user123",
  "recommended_fields": ["Génie Informatique", "Data Science"]
}
```

### 4. PORA returns universities + linked programs
```json
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
```

### 5. Frontend displays results
```
🏫 Université de Kinshasa (Score: 85%)
   → Génie Informatique
   → Data Science
   → Développement Web
```

---

## Error Handling

✅ **Graceful Degradation:**
- If filieres query fails → returns empty array `[]`
- Never crashes the handler
- Logs warning: `⚠️ Error fetching filieres...`
- Universities/centres still returned even if filieres fetch fails

✅ **Empty Results:**
- No universities/centres matched → `univFilieres: []` (empty array)
- No filieres linked to universities → `univFilieres: []` (empty array)
- Frontend can handle empty arrays gracefully

---

## Performance

✅ **Optimizations:**
- Uses database indexes on `universite_id` and `centre_id`
- Efficient join via foreign keys
- DISTINCT eliminates duplicates
- GO `map[string]bool` deduplication is O(n)
- Limit 1000 prevents excessive payload

⚠️ **Note:** If many universities are returned (~50+), this may fetch many filieres. Consider pagination in future.

---

## Testing

### Test Case 1: Normal Flow
```
1. POST /recommendations/universites with fields
2. Verify response has both "universites" and "univFilieres"
3. Verify univFilieres is non-empty array of strings
```

### Test Case 2: No Fields
```
1. POST /recommendations/universites with empty fields
2. Verify response: { "universites": [], "univFilieres": [] }
```

### Test Case 3: No Linked Filieres
```
1. POST with fields that match universities
2. But those universities have no linked filieres
3. Verify response: { "universites": [...], "univFilieres": [] }
```

---

## Files Modified

1. **supabase.go**
   - Added `fetchFilieresForUniversites()`
   - Added `fetchFilieresForCentres()`

2. **handlers.go**
   - Modified `PostUniversiteRecommendations()`
   - Modified `PostCentreRecommendations()`
   - Updated error responses to include empty arrays

---

## Frontend Integration

### Expected Response Structure
```typescript
interface RecommendationResponse {
  universites: Universite[];
  univFilieres?: string[];
  centres?: Centre[];
  centreFilieres?: string[];
}
```

### Display Logic
```javascript
// For universities
universites.forEach(uni => {
  console.log(`${uni.nom}`);
  univFilieres.forEach(filiere => {
    console.log(`  → ${filiere}`);
  });
});
```

---

## Troubleshooting

### Issue: `univFilieres` is empty but universities are returned

**Possible causes:**
1. No filiere-university links exist in database
2. The linked filieres don't have `nom` field filled
3. Query timeout (though unlikely with limit 1000)

**Debug:**
```
Check logs for: 📚 Filières universités trouvées: [list]
If empty, verify universite_filieres table has data
```

### Issue: API returns 500 error

**Possible causes:**
1. `universite_filieres` or `centres_formation_filieres` table doesn't exist
2. Supabase API key missing or invalid
3. Invalid filter syntax

**Debug:**
```
Check Supabase logs for REST API errors
Verify table names in Supabase console
Test query directly in Supabase
```

---

## Future Enhancements

1. **Pagination** - If many results, implement cursor-based pagination
2. **Filtering** - Allow frontend to request specific filieres only
3. **Ranking** - Sort filieres by relevance score (not just name)
4. **Caching** - Cache `universite_filieres` joins to reduce queries
5. **Stats** - Track which filiere combinations are most popular

---

## References

- See `FILIERES_QUERIES.sql` for detailed SQL examples
- Database schema defined in `add_filieres_relations.sql`
