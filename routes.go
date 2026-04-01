package main

import "github.com/gin-gonic/gin"

/*
Routes HTTP du microservice PORA

Responsabilités :
- Signaux utilisateurs (follow / engagement)
- Lecture publique des scores PORA
- Déclenchement interne du ranking

AUCUN calcul d’orientation ici
*/

func RegisterRoutes(r *gin.Engine) {

	// ------------------------------------------------
	// 📌 ACTIONS UTILISATEURS (SIGNAUX PORA)
	// ------------------------------------------------
	r.POST("/universites/:id/follow", FollowUniversite)
	r.POST("/universites/:id/engage", EngageUniversite)

	r.POST("/centres/:id/follow", FollowCentreFormation)
	r.POST("/centres/:id/engage", EngageCentreFormation)

	// ------------------------------------------------
	// 📌 LISTES DES ENTITÉS (PUBLIC)
	// ------------------------------------------------
	r.GET("/universites", UniversiteList)
	r.GET("/centres", CentreFormationList)

	// ------------------------------------------------
	// 🔐 PORA — DÉCLENCHEMENT DU RANKING (INTERNE)
	// ------------------------------------------------
	pora := r.Group("/pora", InternalOnly())
	{
		pora.POST("/run/universites", RunRankingUniversitesHandler)
		pora.POST("/run/centres", RunRankingCentresHandler)
	}

	// ------------------------------------------------
	// 📌 PORA — LECTURE DES SCORES (PUBLIC)
	// ------------------------------------------------
	r.GET("/pora/universite/:id", GetUniversiteScore)
	r.GET("/pora/centre/:id", GetCentreScore)

	r.GET("/ranking/global", GetGlobalRanking)

	r.GET("/ranking/universites", GetRankUniversites)
	r.GET("/ranking/centres", GetRankCentres)

	// ------------------------------------------------
	// 📌 RECOMMANDATIONS (PROA → PORA)
	// ------------------------------------------------
	r.GET("/recommendations/universites", GetUniversiteRecommendations)
	r.POST("/recommendations/universites", PostUniversiteRecommendations)
	r.GET("/recommendations/centres", GetCentreRecommendations)
	r.POST("/recommendations/centres", PostCentreRecommendations)

}

// ------------------------------------------------------------
// 📌 SCORE PORA D’UNE UNIVERSITÉ
// ------------------------------------------------------------
func GetUniversiteScore(c *gin.Context) {
	id := c.Param("id")

	u, err := fetchUniversiteByID(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if u == nil {
		c.JSON(404, gin.H{"error": "université non trouvée"})
		return
	}

	c.JSON(200, gin.H{
		"id":         u.ID,
		"nom":        u.Nom,
		"score_pora": u.ScorePora,
	})
}

// ------------------------------------------------------------
// 📌 SCORE PORA D’UN CENTRE DE FORMATION
// ------------------------------------------------------------
func GetCentreScore(c *gin.Context) {
	id := c.Param("id")

	centre, err := fetchCentreFormationByID(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if centre == nil {
		c.JSON(404, gin.H{"error": "centre de formation non trouvé"})
		return
	}

	c.JSON(200, gin.H{
		"id":         centre.ID,
		"nom":        centre.Nom,
		"score_pora": centre.ScorePora,
	})
}
