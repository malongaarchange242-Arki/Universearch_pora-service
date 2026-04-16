package main

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/*
Ce fichier gère UNIQUEMENT des actions utilisateurs.
Aucune intelligence métier, aucun ranking ici.

Les signaux produits ici sont consommés plus tard
par l’algorithme PORA.
*/

// ====================================================
// 🔵 UNIVERSITÉS
// ====================================================

// ----------------------------------------------------
// 📌 1. FOLLOW D’UNE UNIVERSITÉ
// Signal de popularité sociale (PORA)
// ----------------------------------------------------
func FollowUniversite(c *gin.Context) {
	uniID := c.Param("id")
	userID := c.GetHeader("x-user-id")

	if uniID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing universite id or x-user-id"})
		return
	}

	payload := gin.H{
		"user_id":       userID,
		"universite_id": uniID,
	}

	u, err := url.Parse(SupabaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid supabase url"})
		return
	}
	u.Path = "/rest/v1/followers_universites"

	_, err = httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "resolution=merge-duplicates").
		SetBody(payload).
		Post(u.String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "followed"})
}

// ----------------------------------------------------
// 📌 2. ENGAGEMENT UTILISATEUR AVEC UNE UNIVERSITÉ
// Signal fort (PORA)
// ----------------------------------------------------
func EngageUniversite(c *gin.Context) {
	uniID := c.Param("id")

	var body struct {
		Type   string `json:"type"`
		UserID string `json:"user_id"`
		PostID string `json:"post_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if uniID == "" || body.Type == "" || body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing universite id, type or user_id"})
		return
	}

	payload := gin.H{
		"universite_id": uniID,
		"type":          body.Type,
		"user_id":       body.UserID,
		"post_id":       body.PostID,
	}

	u, err := url.Parse(SupabaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid supabase url"})
		return
	}
	u.Path = "/rest/v1/engagements_universites"

	_, err = httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(u.String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "engagement added"})
}

// ----------------------------------------------------
// 📌 3. LISTE DES UNIVERSITÉS
// ----------------------------------------------------
func UniversiteList(c *gin.Context) {
	unis, err := fetchUniversites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, unis)
}

// ====================================================
// 🟢 CENTRES DE FORMATION
// ====================================================

// ----------------------------------------------------
// 📌 4. FOLLOW D’UN CENTRE
// ----------------------------------------------------
func FollowCentreFormation(c *gin.Context) {
	centreID := c.Param("id")
	userID := c.GetHeader("x-user-id")

	if centreID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing centre id or x-user-id"})
		return
	}

	payload := gin.H{
		"user_id":   userID,
		"centre_id": centreID,
	}

	u, err := url.Parse(SupabaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid supabase url"})
		return
	}
	u.Path = "/rest/v1/followers_centres_formation"

	_, err = httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "resolution=merge-duplicates").
		SetBody(payload).
		Post(u.String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "followed"})
}

// ----------------------------------------------------
// 📌 5. ENGAGEMENT UTILISATEUR AVEC UN CENTRE
// ----------------------------------------------------
func EngageCentreFormation(c *gin.Context) {
	centreID := c.Param("id")

	var body struct {
		Type   string `json:"type"`
		UserID string `json:"user_id"`
		PostID string `json:"post_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if centreID == "" || body.Type == "" || body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing centre id, type or user_id"})
		return
	}

	payload := gin.H{
		"centre_id": centreID,
		"type":      body.Type,
		"user_id":   body.UserID,
		"post_id":   body.PostID,
	}

	u, err := url.Parse(SupabaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid supabase url"})
		return
	}
	u.Path = "/rest/v1/engagements_centres_formation"

	_, err = httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(u.String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "engagement added"})
}

// ----------------------------------------------------
// 📌 6. LISTE DES CENTRES
// ----------------------------------------------------
func CentreFormationList(c *gin.Context) {
	centres, err := fetchCentresFormation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, centres)
}

// ====================================================
// 🔴 RANKING PORA (INTERNE)
// ====================================================

// ----------------------------------------------------
// 📌 7. RANKING UNIVERSITÉS
// ----------------------------------------------------
func RunRankingUniversitesHandler(c *gin.Context) {
	if err := RunRanking(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "PORA universités recalculé"})
}

// ----------------------------------------------------
// 📌 8. RANKING CENTRES
// ----------------------------------------------------
func RunRankingCentresHandler(c *gin.Context) {
	if err := RunRankingCentres(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "PORA centres recalculé"})
}

func GetGlobalRanking(c *gin.Context) {
	nodes, err := RankGlobal()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, nodes)
}

func GetRankUniversites(c *gin.Context) {
	userID := c.Query("user_id")

	var result interface{}
	var err error

	if userID != "" {
		// ✅ Ranking personnalisé pour l'utilisateur
		log.Printf("📊 Ranking enrichi pour user_id=%s", userID)
		result, err = GetRankUniversitesForUser(userID)
	} else {
		// ⚪ Ranking global (pas de profil utilisateur)
		nodes, err := RankUniversites()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for i := range nodes {
			if nodes[i].Type == "universite" {
				u, err := fetchUniversiteByID(nodes[i].ID)
				if err == nil && u != nil {
					nodes[i].Nom = u.Nom
				}
			}
		}
		result = nodes
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

func GetRankCentres(c *gin.Context) {
	data, err := RankCentresFormation()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

type UniversiteRecommendation struct {
	UniversiteID string  `json:"universite_id"`
	NbUsers      int     `json:"nb_users"`
	ScoreMoyen   float64 `json:"score_moyen"`
	ScoreTotal   float64 `json:"score_total"`
}

type CentreRecommendation struct {
	CentreID   string  `json:"centre_id"`
	NbUsers    int     `json:"nb_users"`
	ScoreMoyen float64 `json:"score_moyen"`
	ScoreTotal float64 `json:"score_total"`
}

func GetUniversiteRecommendations(c *gin.Context) {

	data, err := fetchUniversiteRecommendationView()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func PostUniversiteRecommendations(c *gin.Context) {
	// 📥 Recevoir les filières recommandées de PROA + profile_id pour traçabilité
	var body struct {
		UserID            string             `json:"user_id"`
		ProfileID         string             `json:"profile_id"` // 🔗 NOUVEAU: Traçabilité vers PROA
		RecommendedFields []string           `json:"recommended_fields"`
		FieldScores       map[string]float64 `json:"field_scores,omitempty"`
		QuizType          string             `json:"quiz_type"`
		UserType          string             `json:"user_type"` // 🔐 From JWT (frontend extracted)
	}

	// Log le body brut pour debug
	rawBody := c.Request.Body
	defer rawBody.Close()
	bodyBytes, _ := io.ReadAll(rawBody)
	log.Printf("📨 Raw body reçu: %s", string(bodyBytes))

	// Remettre le body pour le parsing
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("❌ POST /recommendations/universites - Bind error: %v", err)
		log.Printf("   Body reçu: %s", string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "details": err.Error()})
		return
	}

	// ✅ Générer UN SEUL session_id pour cette requête
	sessionID := uuid.New().String()
	log.Printf("📊 [SESSION] Nouvelle session créée: %s", sessionID)
	log.Printf("🎯 POST /recommendations/universites - User: %s, Fields: %v", body.UserID, body.RecommendedFields)
	log.Printf("🔥 DEBUG - Field count: %d, Field values: %#v", len(body.RecommendedFields), body.RecommendedFields)
	log.Printf("=== PORA DEBUG ===\nField scores: %#v", body.FieldScores)

	if strings.TrimSpace(body.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	// 🔒 Gestion des filières vides
	if body.RecommendedFields == nil || len(body.RecommendedFields) == 0 {
		if err := replaceOrientationRecommendations(body.UserID, body.ProfileID, "universite", nil, nil, sessionID, map[string]string{}, body.UserType); err != nil {
			log.Printf("❌ POST /recommendations/universites - Persist empty recommendations error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("⚠️ Pas de filières recommandées - retour array vide")
		c.JSON(http.StatusOK, gin.H{
			"universites":  []map[string]interface{}{},
			"univFilieres": []string{},
		})
		return
	}

	// 🔥 Filtrer les universités basées sur les filières recommandées
	filteredUniversites, err := filterUniversitesByFields(body.RecommendedFields)
	if err != nil {
		log.Printf("❌ POST /recommendations/universites - Filter error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ✅ S'assurer que le résultat n'est jamais nil
	if filteredUniversites == nil {
		filteredUniversites = []map[string]interface{}{}
	}

	// 🎓 Récupérer les IDs et noms des universités filtrées
	universiteIDs := make([]string, 0)
	universiteNames := make(map[string]string)
	for _, uni := range filteredUniversites {
		if id, ok := uni["id"].(string); ok {
			universiteIDs = append(universiteIDs, id)
			// Extraire le nom avec fallback
			nom := ""
			if n, ok := uni["nom"].(string); ok && strings.TrimSpace(n) != "" {
				nom = n
			}
			universiteNames[id] = nom
			log.Printf("UNI DEBUG: ID=%s, NOM=%s", id, nom)
		}
	}

	if err := replaceOrientationRecommendations(body.UserID, body.ProfileID, "universite", universiteIDs, body.RecommendedFields, sessionID, universiteNames, body.UserType); err != nil {
		log.Printf("❌ POST /recommendations/universites - Persist error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 📚 Récupérer les filières liées aux universités
	univFilieres := []string{}
	if len(universiteIDs) > 0 {
		filieres, err := fetchFilieresForUniversites(universiteIDs)
		if err == nil {
			univFilieres = filieres
		}
	}

	log.Printf("✅ Universités filtrées: %d résultats", len(filteredUniversites))
	log.Printf("📚 Filières universités: %v", univFilieres)

	log.Printf("🔥 BEFORE ENRICHMENT - universities: %d items", len(filteredUniversites))

	// 🔥 DEBUG SCORING - Ajout du debug demandé
	fmt.Println("---- DEBUG UNIVERSITES ----")
	for _, u := range filteredUniversites {
		name := "Unknown"
		if n, ok := u["nom"].(string); ok {
			name = n
		}
		fmt.Printf("Nom: %s\n", name)

		// Get matched fields count
		matchedCount := 0
		if id, ok := u["id"].(string); ok {
			if matched, err := fetchMatchedFilieresForUniversite(id, body.RecommendedFields); err == nil {
				matchedCount = len(matched)
				fmt.Printf("Matched: %v (%d)\n", matched, matchedCount)
			}
		}

		// Get total fields count
		totalCount := 0
		if id, ok := u["id"].(string); ok {
			if all, err := fetchRealFilieresForUniversite(id); err == nil {
				totalCount = len(all)
				fmt.Printf("Total: %d\n", totalCount)
			}
		}

		// Calculate current score (if any)
		currentScore := 0.0
		if score, ok := u["score_pora"]; ok {
			if s, ok := score.(float64); ok {
				currentScore = s
			}
		}
		fmt.Printf("Score actuel: %.3f\n", currentScore)

		// Calculate what the score SHOULD be
		if len(body.RecommendedFields) > 0 {
			correctScore := float64(matchedCount) / float64(len(body.RecommendedFields))
			fmt.Printf("Score CORRECT (matched/user_fields): %.3f\n", correctScore)
		}

		fmt.Println("---------------------------")
	}
	fmt.Println("---- FIN DEBUG ----")

	enrichedUniversites := enrichUniversiteRecommendationPayload(filteredUniversites, body.RecommendedFields, body.FieldScores, body.UserID)
	log.Printf("🔥 AFTER ENRICHMENT - universities: %d items", len(enrichedUniversites))

	c.JSON(http.StatusOK, gin.H{
		"universites":  enrichedUniversites,
		"univFilieres": univFilieres,
	})
}

func GetCentreRecommendations(c *gin.Context) {

	data, err := fetchCentreRecommendationView()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func PostCentreRecommendations(c *gin.Context) {
	// 📥 Recevoir les filières recommandées de PROA + profile_id pour traçabilité
	var body struct {
		UserID            string             `json:"user_id"`
		ProfileID         string             `json:"profile_id"` // 🔗 NOUVEAU: Traçabilité vers PROA
		RecommendedFields []string           `json:"recommended_fields"`
		FieldScores       map[string]float64 `json:"field_scores,omitempty"`
		QuizType          string             `json:"quiz_type"`
		UserType          string             `json:"user_type"` // 🔐 From JWT (frontend extracted)
	}

	// Log le body brut pour debug
	rawBody := c.Request.Body
	defer rawBody.Close()
	bodyBytes, _ := io.ReadAll(rawBody)
	log.Printf("📨 Raw body reçu (centres): %s", string(bodyBytes))

	// Remettre le body pour le parsing
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("❌ POST /recommendations/centres - Bind error: %v", err)
		log.Printf("   Body reçu: %s", string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "details": err.Error()})
		return
	}

	// ✅ Générer UN SEUL session_id pour cette requête
	sessionID := uuid.New().String()
	log.Printf("📊 [SESSION] Nouvelle session créée: %s", sessionID)
	log.Printf("🎯 POST /recommendations/centres - User: %s, Fields: %v", body.UserID, body.RecommendedFields)
	log.Printf("=== PORA DEBUG ===\nField scores: %#v", body.FieldScores)

	if strings.TrimSpace(body.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	// 🔒 Gestion des filières vides
	if body.RecommendedFields == nil || len(body.RecommendedFields) == 0 {
		if err := replaceOrientationRecommendations(body.UserID, body.ProfileID, "centre", nil, nil, sessionID, map[string]string{}, body.UserType); err != nil {
			log.Printf("❌ POST /recommendations/centres - Persist empty recommendations error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("⚠️ Pas de filières recommandées - retour array vide")
		c.JSON(http.StatusOK, gin.H{
			"centres":        []map[string]interface{}{},
			"centreFilieres": []string{},
		})
		return
	}

	// 🔥 Filtrer les centres basées sur les filières recommandées
	filteredCentres, err := filterCentresByFields(body.RecommendedFields)
	if err != nil {
		log.Printf("❌ POST /recommendations/centres - Filter error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ✅ S'assurer que le résultat n'est jamais nil
	if filteredCentres == nil {
		filteredCentres = []map[string]interface{}{}
	}

	// 🎓 Récupérer les IDs et noms des centres filtrés
	centreIDs := make([]string, 0)
	centreNames := make(map[string]string)
	for _, centre := range filteredCentres {
		if id, ok := centre["id"].(string); ok {
			centreIDs = append(centreIDs, id)
			// Extraire le nom avec fallback
			nom := ""
			if n, ok := centre["nom"].(string); ok && strings.TrimSpace(n) != "" {
				nom = n
			}
			centreNames[id] = nom
			log.Printf("CENTRE DEBUG: ID=%s, NOM=%s", id, nom)
		}
	}

	if err := replaceOrientationRecommendations(body.UserID, body.ProfileID, "centre", centreIDs, body.RecommendedFields, sessionID, centreNames, body.UserType); err != nil {
		log.Printf("❌ POST /recommendations/centres - Persist error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 📚 Récupérer les filières liées aux centres
	centreFilieres := []string{}
	if len(centreIDs) > 0 {
		filieres, err := fetchFilieresForCentres(centreIDs)
		if err == nil {
			centreFilieres = filieres
		}
	}

	log.Printf("✅ Centres filtrés: %d résultats", len(filteredCentres))
	log.Printf("📚 Filières centres: %v", centreFilieres)

	// 🔥 DEBUG SCORING - Ajout du debug demandé pour les centres
	fmt.Println("---- DEBUG CENTRES ----")
	for _, c := range filteredCentres {
		name := "Unknown"
		if n, ok := c["nom"].(string); ok {
			name = n
		}
		fmt.Printf("Nom: %s\n", name)

		// Get matched fields count
		matchedCount := 0
		if id, ok := c["id"].(string); ok {
			if matched, err := fetchMatchedFilieresForCentre(id, body.RecommendedFields); err == nil {
				matchedCount = len(matched)
				fmt.Printf("Matched: %v (%d)\n", matched, matchedCount)
			}
		}

		// Get total fields count
		totalCount := 0
		if id, ok := c["id"].(string); ok {
			if all, err := fetchRealFilieresForCentre(id); err == nil {
				totalCount = len(all)
				fmt.Printf("Total: %d\n", totalCount)
			}
		}

		// Calculate current score (if any)
		currentScore := 0.0
		if score, ok := c["score_pora"]; ok {
			if s, ok := score.(float64); ok {
				currentScore = s
			}
		}
		fmt.Printf("Score actuel: %.3f\n", currentScore)

		// Calculate what the score SHOULD be
		if len(body.RecommendedFields) > 0 {
			correctScore := float64(matchedCount) / float64(len(body.RecommendedFields))
			fmt.Printf("Score CORRECT (matched/user_fields): %.3f\n", correctScore)
		}

		fmt.Println("---------------------------")
	}
	fmt.Println("---- FIN DEBUG CENTRES ----")

	c.JSON(http.StatusOK, gin.H{
		"centres":        enrichCentreRecommendationPayload(filteredCentres, body.RecommendedFields, body.FieldScores, body.UserID),
		"centreFilieres": centreFilieres,
	})
}

func enrichUniversiteRecommendationPayload(items []map[string]interface{}, recommendedFields []string, fieldScores map[string]float64, userID string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))

	log.Printf("🔥 ENRICHING %d universities with fields: %v", len(items), recommendedFields)

	// Précharger popularité et engagement pour normalisation
	popularity := map[string]float64{}
	engagement := map[string]float64{}
	for _, item := range items {
		if id, ok := item["id"].(string); ok && id != "" {
			if v, err := fetchFollowersCount(id); err == nil {
				popularity[id] = v
			} else {
				log.Printf("⚠️ fetchFollowersCount failed for %s: %v", id, err)
			}
			if v, err := fetchEngagementScore(id); err == nil {
				engagement[id] = v
			} else {
				log.Printf("⚠️ fetchEngagementScore failed for %s: %v", id, err)
			}
		}
	}

	normPopularity := normalizeMinMax(popularity)
	normEngagement := normalizeMinMax(engagement)

	for _, item := range items {
		enriched := make(map[string]interface{})
		for key, value := range item {
			enriched[key] = value
		}

		univID, _ := item["id"].(string)
		if univID == "" {
			continue
		}

		if nom, ok := item["nom"].(string); ok {
			enriched["target_name"] = nom
		} else {
			enriched["target_name"] = "Université"
		}

		enriched["universite_id"] = univID
		enriched["recommended_fields_count"] = len(recommendedFields)

		matchedFields, matchedErr := fetchMatchedFilieresForUniversite(univID, recommendedFields)
		if matchedErr != nil {
			log.Printf("⚠️ Could not fetch matched fields for university %s: %v", univID, matchedErr)
		}
		realFields, realErr := fetchRealFilieresForUniversite(univID)
		if realErr != nil {
			log.Printf("⚠️ Could not fetch real fields for university %s: %v", univID, realErr)
		}
		enriched["matched_fields"] = matchedFields
		enriched["real_fields"] = realFields

		matchedCount := len(matchedFields)
		totalCount := len(realFields)

		// Limit to top 20 most relevant filières to prevent large universities from dominating
		maxFieldsConsidered := 20
		if len(recommendedFields) > maxFieldsConsidered {
			recommendedFields = recommendedFields[:maxFieldsConsidered]
		}

		matchRatio := 0.0
		if totalCount > 0 {
			matchRatio = float64(matchedCount) / float64(totalCount)
		}

		specializationBonus := 0.0
		if matchedCount >= 3 && totalCount < 15 {
			specializationBonus = 0.2
		}

		exactBonus := 0.0
		for _, recField := range recommendedFields {
			for _, uniField := range realFields {
				if normalizeField(recField) == normalizeField(uniField) {
					exactBonus = 0.2 // Increased from 0.15 for better differentiation
					break
				}
			}
			if exactBonus > 0 {
				break
			}
		}

		generalistPenalty := 0.0
		if totalCount > 0 {
			generalistPenalty = math.Log(float64(totalCount)+1.0) * 0.1
			if generalistPenalty > 0.3 {
				generalistPenalty = 0.3
			}
		}

		score := 0.6*matchRatio + 0.2*normPopularity[univID] + 0.1*normEngagement[univID] + 0.1*specializationBonus + exactBonus - generalistPenalty
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}

		// Apply non-linear scaling for better UX differentiation
		score = 1 - math.Exp(-score*3)

		enriched["pora_score"] = score
		enriched["score"] = score
		confidence := 0.7 + (score * 0.3)
		enriched["confidence"] = confidence
		enriched["reason"] = buildOrientationRecommendationReason(recommendedFields)

		if matchedErr == nil {
			log.Printf("🎯 University %s matched fields: %v", univID, matchedFields)
		}
		if realErr == nil {
			log.Printf("🏫 University %s real fields: %v", univID, realFields)
		}
		log.Printf("=== PORA DEBUG === University %s | matchRatio=%.3f | popularity=%.3f | engagement=%.3f | specializationBonus=%.2f | exactBonus=%.2f | penalty=%.2f | final=%.3f",
			univID, matchRatio, normPopularity[univID], normEngagement[univID], specializationBonus, exactBonus, generalistPenalty, score)

		// Compute other_fields = real_fields - matched_fields
		otherFields := make([]string, 0, len(realFields))
		matchedSet := make(map[string]bool, len(matchedFields))
		for _, f := range matchedFields {
			matchedSet[strings.ToLower(strings.TrimSpace(f))] = true
		}
		for _, f := range realFields {
			if f == "" {
				continue
			}
			if !matchedSet[strings.ToLower(strings.TrimSpace(f))] {
				otherFields = append(otherFields, f)
			}
		}
		enriched["other_fields"] = otherFields

		out = append(out, enriched)
	}

	sort.SliceStable(out, func(i, j int) bool {
		iScore, _ := out[i]["score"].(float64)
		jScore, _ := out[j]["score"].(float64)
		if iScore == jScore {
			return deterministicNoise(userID, fmt.Sprintf("%v", out[i]["universite_id"])) > deterministicNoise(userID, fmt.Sprintf("%v", out[j]["universite_id"]))
		}
		return iScore > jScore
	})

	return out
}

func enrichCentreRecommendationPayload(items []map[string]interface{}, recommendedFields []string, fieldScores map[string]float64, userID string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))

	popularity := map[string]float64{}
	engagement := map[string]float64{}
	for _, item := range items {
		if id, ok := item["id"].(string); ok && id != "" {
			if v, err := fetchFollowersCentreCount(id); err == nil {
				popularity[id] = v
			} else {
				log.Printf("⚠️ fetchFollowersCentreCount failed for %s: %v", id, err)
			}
			if v, err := fetchEngagementCentreScore(id); err == nil {
				engagement[id] = v
			} else {
				log.Printf("⚠️ fetchEngagementCentreScore failed for %s: %v", id, err)
			}
		}
	}

	normPopularity := normalizeMinMax(popularity)
	normEngagement := normalizeMinMax(engagement)

	for _, item := range items {
		enriched := make(map[string]interface{})
		for key, value := range item {
			enriched[key] = value
		}

		centreID, _ := item["id"].(string)
		if centreID == "" {
			continue
		}

		if nom, ok := item["nom"].(string); ok {
			enriched["target_name"] = nom
		} else {
			enriched["target_name"] = "Centre"
		}

		enriched["centre_id"] = centreID
		enriched["recommended_fields_count"] = len(recommendedFields)

		matchedFields, matchedErr := fetchMatchedFilieresForCentre(centreID, recommendedFields)
		if matchedErr != nil {
			log.Printf("⚠️ Could not fetch matched fields for centre %s: %v", centreID, matchedErr)
		}
		realFields, realErr := fetchRealFilieresForCentre(centreID)
		if realErr != nil {
			log.Printf("⚠️ Could not fetch real fields for centre %s: %v", centreID, realErr)
		}
		enriched["matched_fields"] = matchedFields
		enriched["real_fields"] = realFields

		matchedCount := len(matchedFields)
		totalCount := len(realFields)

		// Limit to top 20 most relevant filières to prevent large centres from dominating
		maxFieldsConsidered := 20
		if len(recommendedFields) > maxFieldsConsidered {
			recommendedFields = recommendedFields[:maxFieldsConsidered]
		}

		matchRatio := 0.0
		if totalCount > 0 {
			matchRatio = float64(matchedCount) / float64(totalCount)
		}

		specializationBonus := 0.0
		if matchedCount >= 3 && totalCount < 15 {
			specializationBonus = 0.2
		}

		exactBonus := 0.0
		for _, recField := range recommendedFields {
			for _, centreField := range realFields {
				if normalizeField(recField) == normalizeField(centreField) {
					exactBonus = 0.2 // Increased from 0.15 for better differentiation
					break
				}
			}
			if exactBonus > 0 {
				break
			}
		}

		generalistPenalty := 0.0
		if totalCount > 0 {
			generalistPenalty = math.Log(float64(totalCount)+1.0) * 0.1
			if generalistPenalty > 0.3 {
				generalistPenalty = 0.3
			}
		}

		score := 0.6*matchRatio + 0.2*normPopularity[centreID] + 0.1*normEngagement[centreID] + 0.1*specializationBonus + exactBonus - generalistPenalty
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}

		// Add random factor to differentiate centers with similar content
		randomFactor := rand.Float64() * 0.05
		score += randomFactor

		// Apply non-linear scaling for better UX differentiation
		score = 1 - math.Exp(-score*3)

		enriched["pora_score"] = score
		enriched["score"] = score
		confidence := 0.7 + (score * 0.3)
		enriched["confidence"] = confidence
		enriched["reason"] = buildOrientationRecommendationReason(recommendedFields)

		if matchedErr == nil {
			log.Printf("🎯 Centre %s matched fields: %v", centreID, matchedFields)
		}
		if realErr == nil {
			log.Printf("🏢 Centre %s real fields: %v", centreID, realFields)
		}
		log.Printf("=== PORA DEBUG === Centre %s | matchRatio=%.3f | popularity=%.3f | engagement=%.3f | specializationBonus=%.2f | exactBonus=%.2f | penalty=%.2f | randomFactor=%.3f | final=%.3f",
			centreID, matchRatio, normPopularity[centreID], normEngagement[centreID], specializationBonus, exactBonus, generalistPenalty, randomFactor, score)

		otherFields := make([]string, 0, len(realFields))
		matchedSet := make(map[string]bool, len(matchedFields))
		for _, f := range matchedFields {
			matchedSet[strings.ToLower(strings.TrimSpace(f))] = true
		}
		for _, f := range realFields {
			if f == "" {
				continue
			}
			if !matchedSet[strings.ToLower(strings.TrimSpace(f))] {
				otherFields = append(otherFields, f)
			}
		}
		enriched["other_fields"] = otherFields

		out = append(out, enriched)
	}

	var filteredOut []map[string]interface{}
	for _, centre := range out {
		matchedFields, hasMatched := centre["matched_fields"].([]string)
		if hasMatched && len(matchedFields) > 0 {
			filteredOut = append(filteredOut, centre)
			log.Printf("✅ Centre %v kept (has %d matched fields)", centre["target_name"], len(matchedFields))
		} else {
			log.Printf("⚠️ Centre %v filtered out (no matched fields)", centre["target_name"])
		}
	}

	log.Printf("📊 Centres before filtering: %d, after: %d", len(out), len(filteredOut))

	sort.SliceStable(filteredOut, func(i, j int) bool {
		iScore, _ := filteredOut[i]["score"].(float64)
		jScore, _ := filteredOut[j]["score"].(float64)
		if iScore == jScore {
			return deterministicNoise(userID, fmt.Sprintf("%v", filteredOut[i]["centre_id"])) > deterministicNoise(userID, fmt.Sprintf("%v", filteredOut[j]["centre_id"]))
		}
		return iScore > jScore
	})

	return filteredOut
}

func deterministicNoise(userID, targetID string) float64 {
	h := fnv.New32a()
	h.Write([]byte(userID))
	h.Write([]byte("|"))
	h.Write([]byte(targetID))
	return float64(h.Sum32()%100) / 100.0
}
