//Supabase.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

var httpClient *resty.Client

// ============================================================
// INITIALISATION CLIENT HTTP
// ============================================================
func newHTTPClient() *resty.Client {
	c := resty.New()
	c.SetTimeout(10 * time.Second)
	c.SetRetryCount(2).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)
	return c
}

func init() {
	LoadConfig()
	httpClient = newHTTPClient()
}

// ============================================================
// HELPERS
// ============================================================

// Parse robuste du header Content-Range Supabase
func parseSupabaseCount(resp *resty.Response) (float64, error) {
	if resp == nil {
		return 0, fmt.Errorf("réponse HTTP nulle")
	}

	cr := resp.Header().Get("Content-Range")
	if cr == "" {
		return 0, fmt.Errorf("Content-Range manquant")
	}

	parts := strings.Split(cr, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("Content-Range invalide: %s", cr)
	}

	count, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ============================================================
// BUSINESS MAPPING FUNCTIONS
// ============================================================

// normalizeField normalizes a field name for better matching
func normalizeField(field string) string {
	// Convert to lowercase
	normalized := strings.ToLower(field)

	// Replace common variations
	normalized = strings.ReplaceAll(normalized, "&", "et")
	normalized = strings.ReplaceAll(normalized, "and", "et")
	normalized = strings.ReplaceAll(normalized, "é", "e")
	normalized = strings.ReplaceAll(normalized, "è", "e")
	normalized = strings.ReplaceAll(normalized, "ê", "e")
	normalized = strings.ReplaceAll(normalized, "à", "a")
	normalized = strings.ReplaceAll(normalized, "â", "a")
	normalized = strings.ReplaceAll(normalized, "ô", "o")
	normalized = strings.ReplaceAll(normalized, "û", "u")
	normalized = strings.ReplaceAll(normalized, "ï", "i")
	normalized = strings.ReplaceAll(normalized, "ç", "c")

	// Remove punctuation and extra spaces
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = strings.ReplaceAll(normalized, ";", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.ReplaceAll(normalized, "(", " ")
	normalized = strings.ReplaceAll(normalized, ")", " ")
	normalized = strings.ReplaceAll(normalized, "[", " ")
	normalized = strings.ReplaceAll(normalized, "]", " ")
	normalized = strings.ReplaceAll(normalized, "{", " ")
	normalized = strings.ReplaceAll(normalized, "}", " ")

	// Split into words, clean, and rejoin
	words := strings.Fields(normalized)
	var cleanWords []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 1 { // Keep words longer than 1 char
			cleanWords = append(cleanWords, word)
		}
	}

	return strings.Join(cleanWords, " ")
}

// isFieldMatch checks if two fields match using intelligent keyword matching
// Avoids generic "genie" matches that create false positives.
func isFieldMatch(userField, uniField string) bool {
	userNorm := normalizeField(userField)
	uniNorm := normalizeField(uniField)

	if userNorm == "" || uniNorm == "" {
		return false
	}

	if isExcludedField(uniNorm) || isExcludedField(userNorm) {
		return false
	}

	// Keyword-based matching for common business domains
	businessKeywords := []string{"compta", "comptabilite", "finance", "gestion", "business", "commerce", "marketing", "vente", "trade", "entrepreneur", "management", "audit", "controle"}
	if containsAnyText(userNorm, businessKeywords) {
		return containsAnyText(uniNorm, businessKeywords)
	}

	// IT keywords (no generic "genie" to avoid false positives)
	itKeywords := []string{"informatique", "logiciel", "developpement", "programmation", "data", "science", "intelligence", "artificielle", "ia", "reseau", "reseaux", "telecom", "telecommunication", "securite", "cyber", "systeme", "systemes"}

	// Engineering keywords (non-generic)
	engKeywords := []string{"mecanique", "electrique", "electronique", "civil", "chimie", "chimique", "industriel", "procedes", "geologique", "hydrosystemes", "hydrosysteme", "aeronautique"}
	lawKeywords := []string{"droit", "juridique", "justice", "penal", "affaires", "public", "prive", "politique", "diplomatie", "gouvernance"}
	healthKeywords := []string{"sante", "medecine", "medical", "pharmacie", "infirm", "dentaire", "biomed", "kine", "laboratoire"}
	scienceKeywords := []string{"biologie", "biotechnologie", "physique", "mathematique", "mathematiques", "statistique", "recherche", "science"}
	environmentKeywords := []string{"environnement", "ecologie", "qhse", "agronomie", "agriculture", "agro", "hydro", "geologie", "geophysique"}

	if containsGenieInformatique(userNorm) {
		return containsAnyText(uniNorm, itKeywords)
	}

	if containsAnyText(userNorm, itKeywords) {
		return containsAnyText(uniNorm, itKeywords)
	}

	if containsAnyText(userNorm, engKeywords) {
		return containsAnyText(uniNorm, engKeywords)
	}

	if containsAnyText(userNorm, lawKeywords) {
		return containsAnyText(uniNorm, lawKeywords)
	}

	if containsAnyText(userNorm, healthKeywords) {
		return containsAnyText(uniNorm, healthKeywords)
	}

	if containsAnyText(userNorm, scienceKeywords) {
		return containsAnyText(uniNorm, scienceKeywords)
	}

	if containsAnyText(userNorm, environmentKeywords) {
		return containsAnyText(uniNorm, environmentKeywords)
	}

	// Direct substring match (case insensitive), guarded against generic "genie" only
	if strings.Contains(userNorm, uniNorm) || strings.Contains(uniNorm, userNorm) {
		if isGenericGenieOnly(userNorm) || isGenericGenieOnly(uniNorm) {
			return false
		}
		return true
	}

	return false
}

func isExcludedField(text string) bool {
	// Do not exclude whole field families at matching time.
	// The recommendation engine should be able to surface any filiere present in base.
	return false
}

func containsAnyText(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func containsGenieInformatique(text string) bool {
	return strings.Contains(text, "genie informatique") || strings.Contains(text, "ingenierie informatique")
}

func isGenericGenieOnly(text string) bool {
	if !(strings.Contains(text, "genie") || strings.Contains(text, "ingenierie")) {
		return false
	}

	itKeywords := []string{"informatique", "logiciel", "developpement", "programmation", "data", "science", "intelligence", "artificielle", "ia", "reseau", "reseaux", "telecom", "telecommunication", "securite", "cyber", "systeme", "systemes"}
	engKeywords := []string{"mecanique", "electrique", "electronique", "civil", "chimie", "chimique", "industriel", "procedes", "geologique", "hydrosystemes", "hydrosysteme", "aeronautique"}

	return !containsAnyText(text, itKeywords) && !containsAnyText(text, engKeywords)
}

// containsKeyword checks if a slice of words contains a keyword
func containsKeyword(words []string, keyword string) bool {
	for _, word := range words {
		if strings.Contains(word, keyword) {
			return true
		}
	}
	return false
}

// expandRecommendedFields maps PROA recommended fields to actual centre filiere names
// This handles the mismatch between academic field names and vocational training names
func expandRecommendedFields(recommendedFields []string) []string {
	// 🔥 BUSINESS MAPPING: PROA fields -> Centre/University filiere names
	mapping := map[string][]string{
		// IT fields
		"Génie Informatique":                      {"Informaticien, programmeur", "Développement informatique", "Technicien en informatique", "Génie Informatique", "Informatique"},
		"Développement Informatique":              {"Informaticien, programmeur", "Développement informatique", "Génie Informatique", "Informatique"},
		"Architecture des Systèmes Informatiques": {"Informaticien, programmeur", "Technicien en informatique", "Génie Informatique", "Informatique"},
		"Data Science":                            {"Informaticien, programmeur", "Statistiques et analyse de données", "Data Science", "Intelligence Artificielle"},
		"Intelligence Artificielle":               {"Informaticien, programmeur", "Intelligence Artificielle", "Data Science"},

		// Engineering fields
		"Génie Civil":      {"Technicien en bâtiment", "Dessinateur en bâtiment", "Génie Civil"},
		"Génie Électrique": {"Électrotechnicien", "Technicien en électronique", "Génie Électrique"},
		"Génie Mécanique":  {"Technicien en mécanique", "Mécanicien", "Génie Mécanique"},

		// Business fields - MAPPING CRUCIAL POUR ESTAM
		"Comptabilité & Gestion d'Entreprise": {"Comptabilité", "Gestion financière", "Audit", "Comptabilité et Contrôle de Gestion", "Comptabilité & Finances", "Finance", "Gestion", "Économie"},
		"Commerce International":              {"Assistant commercial", "Commerce", "Commerce International", "Marketing"},
		"Marketing":                           {"Assistant commercial", "Communication", "Marketing"},
		"Finance":                             {"Comptabilité", "Gestion financière", "Finance", "Audit"},
		"Économie":                            {"Comptabilité", "Gestion", "Économie", "Finance"},

		// Other fields
		"Droit":         {"Assistant juridique", "Droit"},
		"Médecine":      {"Auxiliaire de santé", "Infirmier", "Médecine"},
		"Pharmacie":     {"Préparateur en pharmacie", "Pharmacie"},
		"Biologie":      {"Technicien de laboratoire", "Biotechnologie", "Biologie"},
		"Chimie":        {"Technicien de laboratoire", "Chimie"},
		"Physique":      {"Technicien de laboratoire", "Physique"},
		"Lettres":       {"Assistant administratif", "Communication", "Lettres"},
		"Histoire":      {"Assistant administratif", "Histoire"},
		"Géographie":    {"Assistant commercial", "Géographie"},
		"Mathématiques": {"Statistiques et analyse de données", "Mathématiques"},
	}

	var expanded []string

	// Add original fields
	expanded = append(expanded, recommendedFields...)

	// Add mapped fields
	for _, field := range recommendedFields {
		if mapped, exists := mapping[field]; exists {
			expanded = append(expanded, mapped...)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var result []string
	for _, field := range expanded {
		if !seen[field] {
			seen[field] = true
			result = append(result, field)
		}
	}

	return result
}

// ============================================================
// 🔵 UNIVERSITÉS
// ============================================================

func fetchUniversites() ([]Universite, error) {
	var rows []Universite

	u, _ := url.Parse(SupabaseURL + "/rest/v1/universites")
	q := u.Query()
	q.Set("select", "*")
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Range", "0-9999").
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("fetchUniversites HTTP %d", resp.StatusCode())
	}

	log.Printf("✅ %d universités chargées", len(rows))
	return rows, nil
}

func updateUniversiteScore(id string, score float64) error {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/universites")
	q := u.Query()
	q.Set("id", "eq."+id)
	u.RawQuery = q.Encode()

	payload := map[string]float64{"score_pora": score}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "return=minimal").
		SetBody(payload).
		Patch(u.String())

	if err != nil || resp.IsError() {
		status := 0
		body := ""
		if resp != nil {
			status = resp.StatusCode()
			body = resp.String()
		}
		return fmt.Errorf("update universite %s HTTP %d | body=%s", id, status, body)
	}

	return nil
}

// FOLLOWERS
func fetchFollowersCount(universiteID string) (float64, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/followers_universites")
	q := u.Query()
	q.Set("universite_id", "eq."+universiteID)
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Prefer", "count=exact").
		Head(u.String())

	if err != nil || resp.IsError() {
		return 0, fmt.Errorf("followers universite HTTP %d", resp.StatusCode())
	}

	return parseSupabaseCount(resp)
}

// ENGAGEMENTS
func fetchEngagementScore(universiteID string) (float64, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/engagements_universites")
	q := u.Query()
	q.Set("universite_id", "eq."+universiteID)
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Prefer", "count=exact").
		Head(u.String())

	if err != nil || resp.IsError() {
		return 0, fmt.Errorf("engagement universite HTTP %d", resp.StatusCode())
	}

	return parseSupabaseCount(resp)
}

// ============================================================
// 🟢 CENTRES DE FORMATION
// ============================================================

func fetchCentresFormation() ([]CentreFormation, error) {
	var rows []CentreFormation

	u, _ := url.Parse(SupabaseURL + "/rest/v1/centres_formation")
	q := u.Query()
	q.Set("select", "*")
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Range", "0-9999").
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("fetch centres HTTP %d", resp.StatusCode())
	}

	log.Printf("✅ %d centres chargés", len(rows))
	return rows, nil
}

func updateCentreFormationScore(id string, score float64) error {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/centres_formation")
	q := u.Query()
	q.Set("id", "eq."+id)
	u.RawQuery = q.Encode()

	payload := map[string]float64{"score_pora": score}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "return=minimal").
		SetBody(payload).
		Patch(u.String())

	if err != nil || resp.IsError() {
		return fmt.Errorf("update centre %s HTTP %d", id, resp.StatusCode())
	}

	return nil
}

func fetchFollowersCentreCount(id string) (float64, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/followers_centres_formation")
	q := u.Query()
	q.Set("centre_id", "eq."+id)
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Prefer", "count=exact").
		Head(u.String())

	if err != nil || resp.IsError() {
		return 0, fmt.Errorf("followers centre HTTP %d", resp.StatusCode())
	}

	return parseSupabaseCount(resp)
}

func fetchEngagementCentreScore(id string) (float64, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/engagements_centres_formation")
	q := u.Query()
	q.Set("centre_id", "eq."+id)
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Prefer", "count=exact").
		Head(u.String())

	if err != nil || resp.IsError() {
		return 0, fmt.Errorf("engagement centre HTTP %d", resp.StatusCode())
	}

	return parseSupabaseCount(resp)
}

// ============================================================
// 🔁 RECOMMANDATIONS CROISÉES (PORA)
// ============================================================

func fetchCrossRecommendationScores() (map[string]float64, error) {
	return fetchCrossScores("universite", "centre")
}

func fetchCrossRecommendationScoresCentres() (map[string]float64, error) {
	return fetchCrossScores("centre", "universite")
}

func fetchCrossScores(toType, fromType string) (map[string]float64, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/formation_recommandations_cross")
	q := u.Query()
	q.Set("select", "to_id,poids")
	q.Set("to_type", "eq."+toType)
	q.Set("from_type", "eq."+fromType)
	u.RawQuery = q.Encode()

	var rows []struct {
		ToID  string  `json:"to_id"`
		Poids float64 `json:"poids"`
	}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("cross recommandations HTTP %d", resp.StatusCode())
	}

	scores := map[string]float64{}
	for _, r := range rows {
		scores[r.ToID] += r.Poids
	}

	return scores, nil
}

// ============================================================
// 🧠 ORIENTATION — DB ONLY (PROA → PORA)
// ============================================================

func fetchOrientationScoresUniversites() (map[string]float64, error) {
	return fetchOrientationScores("universite")
}

func fetchOrientationScoresCentres() (map[string]float64, error) {
	return fetchOrientationScores("centre")
}

func fetchOrientationScores(targetType string) (map[string]float64, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/orientation_recommendations")
	q := u.Query()
	q.Set("select", "target_id,score")
	q.Set("target_type", "eq."+targetType)
	u.RawQuery = q.Encode()

	var rows []struct {
		TargetID string  `json:"target_id"`
		Score    float64 `json:"score"`
	}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("orientation %s HTTP %d", targetType, resp.StatusCode())
	}

	sum := map[string]float64{}
	count := map[string]int{}

	for _, r := range rows {
		sum[r.TargetID] += r.Score
		count[r.TargetID]++
	}

	out := map[string]float64{}
	for id, s := range sum {
		out[id] = s / float64(count[id])
	}

	return out, nil
}

func replaceOrientationRecommendations(userID, profileID, targetType string, targetIDs []string, recommendedFields []string, sessionID string, targetNames map[string]string, userType string) error {
	userID = strings.TrimSpace(userID)
	profileID = strings.TrimSpace(profileID)
	targetType = strings.TrimSpace(targetType)
	sessionID = strings.TrimSpace(sessionID)

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	if targetType != "universite" && targetType != "centre" {
		return fmt.Errorf("unsupported target_type: %s", targetType)
	}

	if err := ensureUserExists(userID, userType); err != nil {
		return fmt.Errorf("ensure user exists: %w", err)
	}

	if err := deleteOrientationRecommendations(userID, targetType); err != nil {
		return err
	}

	targetIDs = uniqueOrderedIDs(targetIDs)
	if len(targetIDs) == 0 {
		return nil
	}

	reason := buildOrientationRecommendationReason(recommendedFields)
	rows := make([]map[string]interface{}, 0, len(targetIDs))
	minimalRows := make([]map[string]interface{}, 0, len(targetIDs))

	for idx, targetID := range targetIDs {
		score := deriveOrientationRecommendationScore(idx, len(targetIDs))

		// ✅ Extraire le nom avec fallback propre
		targetName := targetNames[targetID]
		if strings.TrimSpace(targetName) == "" {
			if targetType == "universite" {
				targetName = "Université inconnue"
			} else {
				targetName = "Centre inconnue"
			}
		}

		// ✅ Calculer confidence de manière sophistiquée
		confidence := 0.7 + (score * 0.3) // Base 0.7 + variation selon score

		rows = append(rows, map[string]interface{}{
			"user_id":     userID,
			"profile_id":  profileID, // 🔗 Traçabilité vers PROA
			"session_id":  sessionID, // ✅ Groupe toutes les recommandations de cette session
			"target_type": targetType,
			"target_id":   targetID,
			"target_name": targetName, // ✅ FIX: Inclure le nom
			"score":       score,
			"rank":        idx + 1,
			"confidence":  confidence, // ✅ FIX: Calcul sophistiqué
			"reason":      reason,
		})

		minimalRows = append(minimalRows, map[string]interface{}{
			"user_id":     userID,
			"profile_id":  profileID, // 🔗 Traçabilité vers PROA
			"session_id":  sessionID, // ✅ Même en fallback minimal
			"target_type": targetType,
			"target_id":   targetID,
			"target_name": targetName, // ✅ FIX: Inclure le nom
			"score":       score,
			"rank":        idx + 1,    // ✅ FIX: Ajouter rank au fallback
			"confidence":  confidence, // ✅ FIX: Même calcul sophistiqué
		})
	}

	if err := insertOrientationRecommendations(rows); err != nil {
		log.Printf("⚠️ Insert enrichi orientation_recommendations a échoué, fallback minimal: %v", err)
		if fallbackErr := insertOrientationRecommendations(minimalRows); fallbackErr != nil {
			return fallbackErr
		}
	}

	log.Printf("✅ %d recommandations %s persistées - user_id=%s | profile_id=%s | session_id=%s", len(rows), targetType, userID, profileID, sessionID)
	return nil
}

func insertOrientationRecommendations(rows []map[string]interface{}) error {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/orientation_recommendations")
	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "return=minimal").
		SetBody(rows).
		Post(u.String())

	if err != nil {
		return fmt.Errorf("insert orientation recommendations: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("insert orientation recommendations HTTP %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func ensureUserExists(userID, userType string) error {
	userID = strings.TrimSpace(userID)
	userType = normalizeOrientationUserType(userType)
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	u, _ := url.Parse(SupabaseURL + "/rest/v1/utilisateurs")
	q := u.Query()
	q.Set("select", "id")
	q.Set("id", "eq."+userID)
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Accept", "application/json").
		Get(u.String())

	if err != nil {
		return fmt.Errorf("check utilisateur existence: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("check utilisateur HTTP %d: %s", resp.StatusCode(), resp.String())
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &rows); err != nil {
		return fmt.Errorf("parse utilisateur check response: %w", err)
	}

	if len(rows) > 0 {
		return nil
	}

	pu, _ := url.Parse(SupabaseURL + "/rest/v1/profiles")
	pq := pu.Query()
	pq.Set("select", "id,email")
	pq.Set("id", "eq."+userID)
	pu.RawQuery = pq.Encode()

	presp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Accept", "application/json").
		Get(pu.String())

	if err != nil {
		return fmt.Errorf("check profile existence: %w", err)
	}
	if presp.IsError() {
		return fmt.Errorf("check profile HTTP %d: %s", presp.StatusCode(), presp.String())
	}

	var prows []map[string]interface{}
	if err := json.Unmarshal(presp.Body(), &prows); err != nil {
		return fmt.Errorf("parse profile check response: %w", err)
	}

	if len(prows) == 0 {
		email := ""
		authURL := SupabaseURL + "/auth/v1/admin/users/" + userID
		authResp, err := httpClient.R().
			SetHeader("apikey", SupabaseService).
			SetHeader("Authorization", "Bearer "+SupabaseService).
			SetHeader("Accept", "application/json").
			Get(authURL)

		if err != nil {
			return fmt.Errorf("get auth user: %w", err)
		}

		if authResp.IsError() {
			if authResp.StatusCode() == 404 && strings.Contains(strings.ToLower(authResp.String()), "user_not_found") {
				email = fallbackOrientationProfileEmail(userID)
				log.Printf("ensureUserExists using fallback profile for missing auth user: id=%s | email=%s", userID, email)
			} else {
				return fmt.Errorf("get auth user HTTP %d: %s", authResp.StatusCode(), authResp.String())
			}
		} else {
			var authUser map[string]interface{}
			if err := json.Unmarshal(authResp.Body(), &authUser); err != nil {
				return fmt.Errorf("parse auth user response: %w", err)
			}

			authEmail, ok := authUser["email"].(string)
			if !ok || strings.TrimSpace(authEmail) == "" {
				email = fallbackOrientationProfileEmail(userID)
				log.Printf("ensureUserExists auth user has no email, using fallback: id=%s | email=%s", userID, email)
			} else {
				email = strings.TrimSpace(authEmail)
			}
		}

		profileBody := []map[string]interface{}{{
			"id":           userID,
			"email":        email,
			"profile_type": "utilisateur",
		}}

		pu2, _ := url.Parse(SupabaseURL + "/rest/v1/profiles")
		presp2, err := httpClient.R().
			SetHeader("apikey", SupabaseService).
			SetHeader("Authorization", "Bearer "+SupabaseService).
			SetHeader("Content-Type", "application/json").
			SetHeader("Prefer", "return=minimal").
			SetBody(profileBody).
			Post(pu2.String())

		if err != nil {
			return fmt.Errorf("create profile: %w", err)
		}
		if presp2.IsError() && presp2.StatusCode() != 409 {
			return fmt.Errorf("create profile HTTP %d: %s", presp2.StatusCode(), presp2.String())
		}

		log.Printf("Profile ensured: id=%s | email=%s", userID, email)
	}

	insertBody := []map[string]interface{}{{
		"id":        userID,
		"user_type": userType,
	}}

	u2, _ := url.Parse(SupabaseURL + "/rest/v1/utilisateurs")
	resp2, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "return=minimal").
		SetBody(insertBody).
		Post(u2.String())

	if err != nil {
		return fmt.Errorf("create utilisateur: %w", err)
	}
	if resp2.IsError() && resp2.StatusCode() != 409 {
		return fmt.Errorf("create utilisateur HTTP %d: %s", resp2.StatusCode(), resp2.String())
	}

	log.Printf("User ensured: id=%s | user_type=%s", userID, userType)
	return nil
}

func normalizeOrientationUserType(userType string) string {
	switch strings.TrimSpace(strings.ToLower(userType)) {
	case "bachelier", "etudiant", "parent":
		return strings.TrimSpace(strings.ToLower(userType))
	default:
		return "bachelier"
	}
}

func fallbackOrientationProfileEmail(userID string) string {
	return fmt.Sprintf("pora+%s@universearch.local", strings.ToLower(strings.TrimSpace(userID)))
}

func deleteOrientationRecommendations(userID, targetType string) error {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/orientation_recommendations")
	q := u.Query()
	q.Set("user_id", "eq."+userID)
	q.Set("target_type", "eq."+targetType)
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Prefer", "return=minimal").
		Delete(u.String())

	if err != nil {
		return fmt.Errorf("delete orientation recommendations: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("delete orientation recommendations HTTP %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func deriveOrientationRecommendationScore(index, total int) float64 {
	if total <= 1 {
		return 1
	}

	score := float64(total-index) / float64(total)
	if score < 0.05 {
		return 0.05
	}

	return score
}

func uniqueOrderedIDs(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}

		seen[id] = true
		out = append(out, id)
	}

	return out
}

func buildOrientationRecommendationReason(recommendedFields []string) string {
	cleaned := make([]string, 0, len(recommendedFields))
	seen := make(map[string]bool, len(recommendedFields))

	for _, field := range recommendedFields {
		field = strings.TrimSpace(field)
		if field == "" || seen[field] {
			continue
		}

		seen[field] = true
		cleaned = append(cleaned, field)
		if len(cleaned) == 3 {
			break
		}
	}

	if len(cleaned) == 0 {
		return "Matched recommended fields"
	}

	return "Matched fields: " + strings.Join(cleaned, ", ")
}

func fetchUniversiteByID(id string) (*Universite, error) {
	var rows []Universite

	u, _ := url.Parse(SupabaseURL + "/rest/v1/universites")
	q := u.Query()
	q.Set("id", "eq."+id)
	q.Set("select", "*")
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("fetch universite %s HTTP %d", id, resp.StatusCode())
	}

	if len(rows) == 0 {
		return nil, nil // non trouvée
	}

	return &rows[0], nil
}

func fetchCentreFormationByID(id string) (*CentreFormation, error) {
	var rows []CentreFormation

	u, _ := url.Parse(SupabaseURL + "/rest/v1/centres_formation")
	q := u.Query()
	q.Set("id", "eq."+id)
	q.Set("select", "*")
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("fetch centre %s HTTP %d", id, resp.StatusCode())
	}

	if len(rows) == 0 {
		return nil, nil // non trouvé
	}

	return &rows[0], nil
}

// ============================================================
// 🔵 MISE À JOUR SCORE + SCORE_DETAILS
// ============================================================

func updateUniversiteScoreWithDetails(id string, score float64, details ScoreDetails) error {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/universites")
	q := u.Query()
	q.Set("id", "eq."+id)
	u.RawQuery = q.Encode()

	// Convertir ScoreDetails en JSON pour le payload
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"score_pora":    score,
		"score_details": json.RawMessage(detailsJSON),
	}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "return=minimal").
		SetBody(payload).
		Patch(u.String())

	if err != nil || resp.IsError() {
		status := 0
		body := ""
		if resp != nil {
			status = resp.StatusCode()
			body = resp.String()
		}
		return fmt.Errorf("update universite %s HTTP %d | body=%s", id, status, body)
	}

	return nil
}

func updateCentreScoreWithDetails(id string, score float64, details ScoreDetails) error {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/centres_formation")
	q := u.Query()
	q.Set("id", "eq."+id)
	u.RawQuery = q.Encode()

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"score_pora":    score,
		"score_details": json.RawMessage(detailsJSON),
	}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "return=minimal").
		SetBody(payload).
		Patch(u.String())

	if err != nil || resp.IsError() {
		status := 0
		body := ""
		if resp != nil {
			status = resp.StatusCode()
			body = resp.String()
		}
		return fmt.Errorf("update centre %s HTTP %d | body=%s", id, status, body)
	}

	return nil
}

func fetchUniversiteRecommendationView() ([]UniversiteRecommendation, error) {
	var rows []UniversiteRecommendation

	url := SupabaseURL + "/rest/v1/orientation_scores_universites?select=*&order=score_total.desc"

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(url)

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("error fetching universite recommendations")
	}

	return rows, nil
}

func fetchCentreRecommendationView() ([]CentreRecommendation, error) {
	var rows []CentreRecommendation

	url := SupabaseURL + "/rest/v1/orientation_scores_centres?select=*&order=score_total.desc"

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(url)

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("error fetching centre recommendations")
	}

	return rows, nil
}

// ============================================================
// 🎓 FILIÈRES LINKED TO UNIVERSITÉS
// ============================================================

func fetchFilieresForUniversites(universiteIDs []string) ([]string, error) {
	if len(universiteIDs) == 0 {
		return []string{}, nil
	}

	// Build filter for IN clause
	var inFilter strings.Builder
	for i, id := range universiteIDs {
		if i > 0 {
			inFilter.WriteString(",")
		}
		inFilter.WriteString(fmt.Sprintf("\"%s\"", id))
	}

	u, _ := url.Parse(SupabaseURL + "/rest/v1/universite_filieres")
	q := u.Query()
	q.Set("select", "filieres(nom)")
	q.Set("universite_id", fmt.Sprintf("in.(%s)", inFilter.String()))
	q.Set("limit", "1000")
	u.RawQuery = q.Encode()

	var rows []map[string]interface{}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		log.Printf("⚠️ Error fetching filieres for universites: HTTP %d", resp.StatusCode())
		return []string{}, nil
	}

	// Extract unique filiere names and avoid duplicates
	filieresMap := make(map[string]bool)
	for _, row := range rows {
		if filiere, ok := row["filieres"]; ok {
			if filiere != nil {
				if filiereSub, ok := filiere.(map[string]interface{}); ok {
					if name, ok := filiereSub["nom"].(string); ok {
						filieresMap[name] = true
					}
				}
			}
		}
	}

	// Convert map to slice
	filieresSlice := make([]string, 0, len(filieresMap))
	for name := range filieresMap {
		filieresSlice = append(filieresSlice, name)
	}

	log.Printf("📚 Filières universités trouvées: %v", filieresSlice)
	return filieresSlice, nil
}

// ============================================================
// 🎓 FILIÈRES LINKED TO CENTRES
// ============================================================

func fetchFilieresForCentres(centreIDs []string) ([]string, error) {
	if len(centreIDs) == 0 {
		return []string{}, nil
	}

	// Build filter for IN clause
	var inFilter strings.Builder
	for i, id := range centreIDs {
		if i > 0 {
			inFilter.WriteString(",")
		}
		inFilter.WriteString(fmt.Sprintf("\"%s\"", id))
	}

	u, _ := url.Parse(SupabaseURL + "/rest/v1/centre_formation_filieres")
	q := u.Query()
	q.Set("select", "filieres(nom)")
	q.Set("centre_id", fmt.Sprintf("in.(%s)", inFilter.String()))
	q.Set("limit", "1000")
	u.RawQuery = q.Encode()

	var rows []map[string]interface{}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		log.Printf("⚠️ Error fetching filieres for centres: HTTP %d", resp.StatusCode())
		return []string{}, nil
	}

	// Extract unique filiere names and avoid duplicates
	filieresMap := make(map[string]bool)
	for _, row := range rows {
		if filiere, ok := row["filieres"]; ok {
			if filiere != nil {
				if filiereSub, ok := filiere.(map[string]interface{}); ok {
					if name, ok := filiereSub["nom"].(string); ok {
						filieresMap[name] = true
					}
				}
			}
		}
	}

	// Convert map to slice
	filieresSlice := make([]string, 0, len(filieresMap))
	for name := range filieresMap {
		filieresSlice = append(filieresSlice, name)
	}

	log.Printf("📚 Filières centres trouvées: %v", filieresSlice)
	return filieresSlice, nil
}

// fetchMatchedFilieresForUniversite retourne les filières d'une université qui correspondent aux champs recommandés
func fetchMatchedFilieresForUniversite(univID string, recommendedFields []string) ([]string, error) {
	if univID == "" || len(recommendedFields) == 0 {
		log.Printf("⚠️ fetchMatchedFilieresForUniversite: empty params - univID: %s, fields: %v", univID, recommendedFields)
		return []string{}, nil
	}

	log.Printf("🔍 fetchMatchedFilieresForUniversite: university %s with recommended fields: %v", univID, recommendedFields)

	// First, get all filieres offered by this university
	u, _ := url.Parse(SupabaseURL + "/rest/v1/universite_filieres")
	q := u.Query()
	q.Set("select", "filieres(nom)")
	q.Set("universite_id", "eq."+univID)
	q.Set("limit", "1000")
	u.RawQuery = q.Encode()

	var rows []map[string]interface{}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		log.Printf("⚠️ Error fetching filieres for university %s: HTTP %d", univID, resp.StatusCode())
		return []string{}, err
	}

	// Extract filiere names from this university
	var universityFilieres []string
	for _, row := range rows {
		if filiere, ok := row["filieres"]; ok && filiere != nil {
			if filiereSub, ok := filiere.(map[string]interface{}); ok {
				if name, ok := filiereSub["nom"].(string); ok {
					universityFilieres = append(universityFilieres, name)
				}
			}
		}
	}

	log.Printf("🏫 University %s offers filieres: %v", univID, universityFilieres)

	// 🔥 INTELLIGENT MATCHING: Use new keyword-based matching system
	var matchedFilieres []string

	for _, uniFiliere := range universityFilieres {
		for _, recField := range recommendedFields {
			if isFieldMatch(recField, uniFiliere) {
				matchedFilieres = append(matchedFilieres, uniFiliere)
				log.Printf("✅ Match found: '%s' matches '%s'", recField, uniFiliere)
				break // Found a match for this uni filiere, no need to check other rec fields
			}
		}
	}

	// Hard filter for excluded domains to prevent false positives.
	var filteredMatched []string
	for _, f := range matchedFilieres {
		if !isExcludedField(normalizeField(f)) {
			filteredMatched = append(filteredMatched, f)
		}
	}

	// Remove duplicates and limit to 5
	matchedMap := make(map[string]bool)
	var result []string
	for _, f := range filteredMatched {
		if !matchedMap[f] && len(result) < 5 {
			matchedMap[f] = true
			result = append(result, f)
		}
	}

	log.Printf("✅ Matched filieres for university %s: %v", univID, result)
	return result, nil
}

// fetchMatchedFilieresForCentre retourne les filières d'un centre qui correspondent aux champs recommandés
func fetchMatchedFilieresForCentre(centreID string, recommendedFields []string) ([]string, error) {
	if centreID == "" || len(recommendedFields) == 0 {
		return []string{}, nil
	}

	// 🔥 CORRECTED QUERY: Use proper table pivot centre_formation_filieres
	// First get all filiere_ids for this centre
	u1, _ := url.Parse(SupabaseURL + "/rest/v1/centre_formation_filieres")
	q1 := u1.Query()
	q1.Set("select", "filiere_id")
	q1.Set("centre_id", "eq."+centreID)
	u1.RawQuery = q1.Encode()

	var pivotRows []map[string]interface{}

	resp1, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&pivotRows).
		Get(u1.String())

	if err != nil || resp1.IsError() {
		log.Printf("⚠️ Error fetching pivot data for centre %s: HTTP %d", centreID, resp1.StatusCode())
		return []string{}, nil
	}

	// Extract filiere_ids
	var filiereIDs []string
	for _, row := range pivotRows {
		if fid, ok := row["filiere_id"].(string); ok && fid != "" {
			filiereIDs = append(filiereIDs, fid)
		}
	}

	if len(filiereIDs) == 0 {
		log.Printf("ℹ️ No filieres found for centre %s", centreID)
		return []string{}, nil
	}

	// 🔥 BUSINESS MAPPING: Map PROA recommended fields to actual centre filiere names
	// This handles the mismatch between PROA field names and centre filiere names
	mappedRecommendedFields := expandRecommendedFields(recommendedFields)
	log.Printf("🔄 Mapped recommended fields: %v -> %v", recommendedFields, mappedRecommendedFields)

	// 🔥 INTELLIGENT MATCHING: Get ALL filieres first, then match in Go
	// This allows for sophisticated keyword matching instead of exact string matching

	// Get ALL filiere names for this centre (no filtering)
	u2, _ := url.Parse(SupabaseURL + "/rest/v1/filieres_centre")
	q2 := u2.Query()
	q2.Set("select", "nom")
	q2.Set("id", "in.("+strings.Join(filiereIDs, ",")+")")
	u2.RawQuery = q2.Encode()

	var filiereRows []map[string]interface{}

	resp2, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&filiereRows).
		Get(u2.String())

	if err != nil || resp2.IsError() {
		log.Printf("⚠️ Error fetching filiere names for centre %s: HTTP %d", centreID, resp2.StatusCode())
		return []string{}, nil
	}

	// Extract ALL centre filiere names
	var centreFilieres []string
	for _, row := range filiereRows {
		if name, ok := row["nom"].(string); ok && name != "" {
			centreFilieres = append(centreFilieres, name)
		}
	}

	log.Printf("🏢 Centre %s offers filieres: %v", centreID, centreFilieres)

	// 🔥 INTELLIGENT MATCHING: Use new keyword-based matching system
	var matchedFilieres []string
	for _, centreFiliere := range centreFilieres {
		for _, recField := range mappedRecommendedFields {
			if isFieldMatch(recField, centreFiliere) {
				matchedFilieres = append(matchedFilieres, centreFiliere)
				log.Printf("✅ Match found: '%s' matches '%s'", recField, centreFiliere)
				break // Found a match for this centre filiere
			}
		}
	}

	// Hard filter for excluded domains to prevent false positives.
	var filteredMatched []string
	for _, f := range matchedFilieres {
		if !isExcludedField(normalizeField(f)) {
			filteredMatched = append(filteredMatched, f)
		}
	}

	// Remove duplicates and limit to 5
	matchedMap := make(map[string]bool)
	var result []string
	for _, f := range filteredMatched {
		if !matchedMap[f] && len(result) < 5 {
			matchedMap[f] = true
			result = append(result, f)
		}
	}

	log.Printf("✅ Matched filieres for centre %s: %v", centreID, result)
	return result, nil
}

// fetchRealFilieresForCentre retourne TOUTES les filières réelles proposées par un centre
func fetchRealFilieresForCentre(centreID string) ([]string, error) {
	if centreID == "" {
		return []string{}, nil
	}

	log.Printf("🔍 Fetching ALL real filieres for centre %s", centreID)

	// Get all filiere_ids for this centre
	u1, _ := url.Parse(SupabaseURL + "/rest/v1/centre_formation_filieres")
	q1 := u1.Query()
	q1.Set("select", "filiere_id")
	q1.Set("centre_id", "eq."+centreID)
	u1.RawQuery = q1.Encode()

	var pivotRows []map[string]interface{}

	resp1, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&pivotRows).
		Get(u1.String())

	if err != nil || resp1.IsError() {
		log.Printf("⚠️ Error fetching pivot data for centre %s: HTTP %d", centreID, resp1.StatusCode())
		return []string{}, nil
	}

	// Extract filiere_ids
	var filiereIDs []string
	for _, row := range pivotRows {
		if fid, ok := row["filiere_id"].(string); ok && fid != "" {
			filiereIDs = append(filiereIDs, fid)
		}
	}

	if len(filiereIDs) == 0 {
		log.Printf("ℹ️ No filieres found for centre %s", centreID)
		return []string{}, nil
	}

	// Get ALL filiere names for this centre (no filtering by recommended fields)
	u2, _ := url.Parse(SupabaseURL + "/rest/v1/filieres_centre")
	q2 := u2.Query()
	q2.Set("select", "nom")
	q2.Set("id", "in.("+strings.Join(filiereIDs, ",")+")")
	u2.RawQuery = q2.Encode()

	var filiereRows []map[string]interface{}

	resp2, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&filiereRows).
		Get(u2.String())

	if err != nil || resp2.IsError() {
		log.Printf("⚠️ Error fetching filiere names for centre %s: HTTP %d", centreID, resp2.StatusCode())
		return []string{}, nil
	}

	// Extract ALL real filiere names
	var realFilieres []string
	for _, row := range filiereRows {
		if name, ok := row["nom"].(string); ok && name != "" {
			realFilieres = append(realFilieres, name)
		}
	}

	log.Printf("✅ Real filieres for centre %s: %v", centreID, realFilieres)
	return realFilieres, nil
}

// fetchRealFilieresForUniversite retourne TOUTES les filières réelles proposées par une université
func fetchRealFilieresForUniversite(univID string) ([]string, error) {
	if univID == "" {
		return []string{}, nil
	}

	log.Printf("🔍 Fetching ALL real filieres for university %s", univID)

	// Get all filiere_ids for this university
	u, _ := url.Parse(SupabaseURL + "/rest/v1/universite_filieres")
	q := u.Query()
	q.Set("select", "filieres(nom)")
	q.Set("universite_id", "eq."+univID)
	q.Set("limit", "1000")
	u.RawQuery = q.Encode()

	var rows []map[string]interface{}

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		log.Printf("⚠️ Error fetching filieres for university %s: HTTP %d", univID, resp.StatusCode())
		return []string{}, nil
	}

	// Extract filiere names from this university
	var realFilieres []string
	for _, row := range rows {
		if filiere, ok := row["filieres"]; ok && filiere != nil {
			if filiereSub, ok := filiere.(map[string]interface{}); ok {
				if name, ok := filiereSub["nom"].(string); ok {
					realFilieres = append(realFilieres, name)
				}
			}
		}
	}

	log.Printf("✅ Real filieres for university %s: %v", univID, realFilieres)
	return realFilieres, nil
}
