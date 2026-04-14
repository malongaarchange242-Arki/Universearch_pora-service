package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// ==================================================
// 📊 ENREGISTREMENT DES RECOMMANDATIONS (Data Driver)
// ==================================================

// Structure pour enregistrer une recommandation
type RecommendationRecord struct {
	ID                   string  `json:"id"`
	UserID               string  `json:"user_id"`
	ProfileID            string  `json:"profile_id"`            // 🔗 Lié au profil PROA
	SessionID            string  `json:"session_id"`            // 📌 Regrouper une session
	TargetType           string  `json:"target_type"`           // "universite" ou "filiere"
	TargetID             string  `json:"target_id"`             // ID de la cible
	TargetName           string  `json:"target_name"`           // Nom lisible
	Score                float64 `json:"score"`                 // Score final (0-1)
	Rank                 int     `json:"rank"`                  // Position dans le ranking
	Confidence           float64 `json:"confidence"`            // Confiance du profil
	Reason               string  `json:"reason"`                // Raison: "77% PORA + 23% Orientation"
	RecommendationEngine string  `json:"recommendation_engine"` // "pora_v1"
	CreatedAt            string  `json:"created_at"`
}

// Générer un session ID unique
func GenerateSessionID() string {
	return uuid.New().String()
}

// Enregistrer les recommandations pour une session
func SaveRecommendations(
	userID string,
	profileID string,
	sessionID string,
	enrichedNodes []EnrichedNode,
	userProfile *OrientationComputeResponse,
) error {

	if userID == "" || sessionID == "" {
		return fmt.Errorf("userID or sessionID cannot be empty")
	}

	records := make([]RecommendationRecord, 0)

	for _, node := range enrichedNodes {
		reason := fmt.Sprintf("%.0f%% PORA", node.ScoreRaw*100)
		if userProfile != nil && profileID != "" {
			orientationPart := node.OrientationScore * 100
			poraPart := (node.ScoreRaw / 0.6) * 100 // Inverser la pondération
			reason = fmt.Sprintf("%.0f%% PORA + %.0f%% Orientation", poraPart, orientationPart)
		}

		record := RecommendationRecord{
			ID:                   uuid.New().String(),
			UserID:               userID,
			ProfileID:            profileID,
			SessionID:            sessionID,
			TargetType:           "universite", // TODO: Utiliser le type du node
			TargetID:             node.ID,
			TargetName:           node.Nom,
			Score:                node.ScoreRaw,
			Rank:                 node.Rank,
			Confidence:           getUserConfidence(userProfile),
			Reason:               reason,
			RecommendationEngine: "pora_v1",
			CreatedAt:            time.Now().UTC().Format(time.RFC3339),
		}

		records = append(records, record)
	}

	// 📊 Insérer en base de données
	err := InsertRecommendationRecords(records)
	if err != nil {
		log.Printf("❌ Erreur insertion recommandations: %v", err)
		return err
	}

	log.Printf("✅ %d recommandations enregistrées pour session %s", len(records), sessionID)
	return nil
}

// Insérer les records en Supabase
func InsertRecommendationRecords(records []RecommendationRecord) error {

	if len(records) == 0 {
		return fmt.Errorf("no records to insert")
	}

	// Convertir en JSON pour l'insertion
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	log.Printf("📤 Insertion de %d recommandations", len(records))
	log.Printf("   Données: %s", string(data[:100]))

	// ⚠️ NOTE: Cette fonction utilise le client Supabase existant
	// Elle doit être appelée depuis un contexte avec SupabaseClient initialisé
	// Voir supabase.go pour l'implémentation
	for _, record := range records {
		err := InsertSingleRecommendation(record)
		if err != nil {
			log.Printf("⚠️  Erreur insertion recommandation %s: %v", record.TargetName, err)
			// Continuer avec les autres (pas de stop sur erreur)
		}
	}

	return nil
}

// Insérer une recommandation unique (wrapper)
func InsertSingleRecommendation(record RecommendationRecord) error {
	// Utiliser l'infrastructure Supabase existante
	rows := []map[string]interface{}{
		{
			"user_id":               record.UserID,
			"profile_id":            record.ProfileID, // 🔗 Lié au profil PROA
			"session_id":            record.SessionID, // 📌 Regrouper les sessions
			"target_type":           record.TargetType,
			"target_id":             record.TargetID,
			"target_name":           record.TargetName,
			"score":                 record.Score,
			"rank":                  record.Rank,
			"confidence":            record.Confidence,
			"reason":                record.Reason,
			"recommendation_engine": record.RecommendationEngine,
			"created_at":            record.CreatedAt,
		},
	}

	// Insérer via Supabase (même pattern que insertOrientationRecommendations)
	err := insertOrientationRecommendations(rows)
	if err != nil {
		log.Printf("❌ Erreur insertion recommandation %s: %v", record.TargetName, err)
		return err
	}

	log.Printf("✅ Recommandation enregistrée: %s (rank=%d)", record.TargetName, record.Rank)
	return nil
}

// Helper: Récupérer la confiance utilisateur
func getUserConfidence(userProfile *OrientationComputeResponse) float64 {
	if userProfile == nil {
		return 0.0
	}
	return userProfile.Confidence
}

// Vue pour analyser les recommandations enregistrées
func GetRecommendationAnalytics(userID string) map[string]interface{} {
	// Cette fonction retourne des analytics sur les recommandations passées
	// Elle sera implémentée dans supabase.go

	return map[string]interface{}{
		"user_id": userID,
		"note":    "À implémenter dans supabase.go",
	}
}
