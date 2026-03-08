package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
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

func GetCentreRecommendations(c *gin.Context) {

	data, err := fetchCentreRecommendationView()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}
