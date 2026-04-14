package main

import (
	"fmt"
	"log"
	"math"
)

// ==================================================
// RANKING ENRICHI PAR PROFIL UTILISATEUR (PROA)
// ==================================================

type UserRankingRequest struct {
	UserID string `json:"user_id"`
}

type EnrichedNode struct {
	PORANode
	OrientationScore float64   `json:"orientation_score"`
	OrientationMatch []float64 `json:"orientation_match,omitempty"`
}

// Obtenir le ranking personnalisé pour un utilisateur
func RankUniversitesForUser(userID string) ([]EnrichedNode, error) {
	log.Printf("[RANK USER] 📊 Calcul ranking personnalisé pour %s", userID)

	// 🆔 Générer un session_id unique pour cette demande
	sessionID := GenerateSessionID()
	log.Printf("🔗 Session ID généré: %s", sessionID)

	// 1️⃣ Récupérer le ranking PORA de base
	baseNodes, err := RankUniversites()
	if err != nil {
		return nil, fmt.Errorf("erreur ranking PORA: %w", err)
	}

	// 2️⃣ Appeler PROA pour le profil utilisateur
	userProfile, err := CallOrientationCompute(userID, "1.0", nil)
	if err != nil {
		log.Printf("⚠️  PROA profile fetch failed: %v (fallback ranking global)", err)
		// Fallback: retourner le ranking global
		return convertToEnrichedNodes(baseNodes, nil), nil
	}

	log.Printf("✅ Profil utilisateur reçu: confidence=%.2f | profile_id=%s", userProfile.Confidence, userProfile.ProfileID)

	// 3️⃣ Enrichir le ranking avec le profil utilisateur
	enrichedNodes := enrichNodesWithUserProfile(baseNodes, userProfile)

	// 4️⃣ Trier par score enrichi
	sortNodesByScore(enrichedNodes)

	// 📊 NOUVEAU: Enregistrer les recommandations pour analyse
	go func() {
		err := SaveRecommendations(userID, userProfile.ProfileID, sessionID, enrichedNodes, userProfile)
		if err != nil {
			log.Printf("⚠️  Erreur enregistrement recommandations: %v", err)
		}
	}()

	return enrichedNodes, nil
}

// Enrichir les nodes avec le score d'orientation utilisateur
func enrichNodesWithUserProfile(
	baseNodes []PORANode,
	userProfile *OrientationComputeResponse,
) []EnrichedNode {

	enriched := make([]EnrichedNode, len(baseNodes))

	for i, node := range baseNodes {
		enriched[i] = EnrichedNode{
			PORANode: node,
		}

		// Si pas de profil, juste copier
		if userProfile == nil {
			enriched[i].OrientationScore = 0
			continue
		}

		// Score basique = moyenne du profil (0-1)
		avgScore := 0.0
		if len(userProfile.Profile) > 0 {
			for _, v := range userProfile.Profile {
				avgScore += v
			}
			avgScore /= float64(len(userProfile.Profile))
		}

		enriched[i].OrientationMatch = userProfile.Profile
		enriched[i].OrientationScore = avgScore

		// Pondération: 60% PORA + 40% Orientation
		enriched[i].ScoreRaw = (enriched[i].ScoreRaw * 0.6) + (avgScore * 0.4)
	}

	return enriched
}

// Convertir les PORANode en EnrichedNode (sans profil)
func convertToEnrichedNodes(baseNodes []PORANode, userProfile *OrientationComputeResponse) []EnrichedNode {
	enriched := make([]EnrichedNode, len(baseNodes))

	for i, node := range baseNodes {
		enriched[i] = EnrichedNode{
			PORANode:         node,
			OrientationScore: 0,
		}
	}

	return enriched
}

// Trier les nodes enrichis par score
func sortNodesByScore(nodes []EnrichedNode) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[j].ScoreRaw > nodes[i].ScoreRaw {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}

	// Appliquer rank et percentile
	total := len(nodes)
	for i := range nodes {
		nodes[i].Rank = i + 1
		if total == 1 {
			nodes[i].Percentile = 100
		} else {
			nodes[i].Percentile = int(math.Round(float64(total-i) * 100 / float64(total)))
		}
	}
}

// Handler HTTP pour le ranking utilisateur
func GetRankUniversitesForUser(userID string) (interface{}, error) {
	enriched, err := RankUniversitesForUser(userID)
	if err != nil {
		return nil, err
	}

	// Normaliser les scores enrichis
	baseScores := make([]PORANode, len(enriched))
	for i, e := range enriched {
		baseScores[i] = e.PORANode
	}
	normalizeByRawScore(baseScores)

	for i := range enriched {
		enriched[i].PORANode = baseScores[i]
	}

	return enriched, nil
}
