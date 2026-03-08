// PORA.go
package main

import (
	"log"
	"math"
)

/*
PORA = Profile-Oriented Recommendation Aggregation
Aucune intelligence artificielle.
Uniquement l’agrégation NORMALISÉE de signaux utilisateurs réels :
- Popularité sociale (followers)
- Activité / engagement réel
- Décisions utilisateurs (orientation PROA)
- Recommandations croisées (universités <-> centres)
*/

// ------------------------------------------------------------
// POIDS PORA (DOIVENT TOUJOURS FAIRE 1.0)
// ------------------------------------------------------------
const (
	WEIGHT_FOLLOWERS   = 0.25
	WEIGHT_ENGAGEMENTS = 0.25
	WEIGHT_ORIENTATION = 0.25
	WEIGHT_CROSS       = 0.25
)

// ============================================================
// 🔵 RUN RANKING — UNIVERSITÉS
// ============================================================
func RunRanking() error {
	log.Println("[PORA] ▶ Classement UNIVERSITÉS")

	universites, err := fetchUniversites()
	if err != nil {
		return err
	}

	if len(universites) == 0 {
		log.Println("[PORA] ℹ️ aucune université à scorer")
		return nil
	}

	followCounts := map[string]float64{}
	engagementScores := map[string]float64{}

	for _, u := range universites {
		followCounts[u.ID] = 0
		engagementScores[u.ID] = 0
	}

	for _, u := range universites {
		if v, err := fetchFollowersCount(u.ID); err == nil {
			followCounts[u.ID] = v
		}
		if v, err := fetchEngagementScore(u.ID); err == nil {
			engagementScores[u.ID] = v
		}
	}

	orientationScores, err := fetchOrientationScoresUniversites()
	if err != nil {
		return err
	}

	crossScores, err := fetchCrossRecommendationScores()
	if err != nil {
		return err
	}

	for _, u := range universites {
		if _, ok := orientationScores[u.ID]; !ok {
			orientationScores[u.ID] = 0
		}
		if _, ok := crossScores[u.ID]; !ok {
			crossScores[u.ID] = 0
		}
	}

	normFollowers := normalizeMinMax(followCounts)
	normEngagements := normalizeMinMax(engagementScores)
	normOrientation := normalizeMinMax(orientationScores)
	normCross := normalizeMinMax(crossScores)

	for _, u := range universites {

		score := normFollowers[u.ID]*WEIGHT_FOLLOWERS +
			normEngagements[u.ID]*WEIGHT_ENGAGEMENTS +
			normOrientation[u.ID]*WEIGHT_ORIENTATION +
			normCross[u.ID]*WEIGHT_CROSS

		details := ScoreDetails{
			Followers:   normFollowers[u.ID],
			Engagement:  normEngagements[u.ID],
			Orientation: normOrientation[u.ID],
			Cross:       normCross[u.ID],
		}

		log.Printf(
			"[PORA][UNI] %s → F=%.2f E=%.2f O=%.2f C=%.2f | SCORE=%.3f",
			u.ID,
			details.Followers,
			details.Engagement,
			details.Orientation,
			details.Cross,
			score,
		)

		if err := updateUniversiteScoreWithDetails(u.ID, score, details); err != nil {
			log.Printf("[PORA] ❌ update université %s: %v", u.ID, err)
		}
	}

	log.Println("[PORA] ✅ Classement UNIVERSITÉS terminé")
	return nil
}

// ============================================================
// 🟢 RUN RANKING — CENTRES DE FORMATION
// ============================================================
func RunRankingCentres() error {
	log.Println("[PORA] ▶ Classement CENTRES DE FORMATION")

	centres, err := fetchCentresFormation()
	if err != nil {
		return err
	}

	if len(centres) == 0 {
		log.Println("[PORA] ℹ️ aucun centre à scorer")
		return nil
	}

	followCounts := map[string]float64{}
	engagementScores := map[string]float64{}

	for _, c := range centres {
		followCounts[c.ID] = 0
		engagementScores[c.ID] = 0
	}

	for _, c := range centres {
		if v, err := fetchFollowersCentreCount(c.ID); err == nil {
			followCounts[c.ID] = v
		}
		if v, err := fetchEngagementCentreScore(c.ID); err == nil {
			engagementScores[c.ID] = v
		}
	}

	orientationScores, err := fetchOrientationScoresCentres()
	if err != nil {
		return err
	}

	crossScores, err := fetchCrossRecommendationScoresCentres()
	if err != nil {
		return err
	}

	for _, c := range centres {
		if _, ok := orientationScores[c.ID]; !ok {
			orientationScores[c.ID] = 0
		}
		if _, ok := crossScores[c.ID]; !ok {
			crossScores[c.ID] = 0
		}
	}

	normFollowers := normalizeMinMax(followCounts)
	normEngagements := normalizeMinMax(engagementScores)
	normOrientation := normalizeMinMax(orientationScores)
	normCross := normalizeMinMax(crossScores)

	for _, c := range centres {

		score := normFollowers[c.ID]*WEIGHT_FOLLOWERS +
			normEngagements[c.ID]*WEIGHT_ENGAGEMENTS +
			normOrientation[c.ID]*WEIGHT_ORIENTATION +
			normCross[c.ID]*WEIGHT_CROSS

		details := ScoreDetails{
			Followers:   normFollowers[c.ID],
			Engagement:  normEngagements[c.ID],
			Orientation: normOrientation[c.ID],
			Cross:       normCross[c.ID],
		}

		log.Printf(
			"[PORA][CENTRE] %s → F=%.2f E=%.2f O=%.2f C=%.2f | SCORE=%.3f",
			c.ID,
			details.Followers,
			details.Engagement,
			details.Orientation,
			details.Cross,
			score,
		)

		if err := updateCentreScoreWithDetails(c.ID, score, details); err != nil {
			log.Printf("[PORA] ❌ update centre %s: %v", c.ID, err)
		}
	}

	log.Println("[PORA] ✅ Classement CENTRES terminé")
	return nil
}

// ------------------------------------------------------------
// NORMALISATION MIN–MAX (ROBUSTE)
// ------------------------------------------------------------
func normalizeMinMax(values map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(values))
	if len(values) == 0 {
		return out
	}

	min := math.MaxFloat64
	max := -math.MaxFloat64

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	for id, v := range values {
		if max == min {
			out[id] = 0.5 // signal neutre si tout est identique
		} else {
			out[id] = (v - min) / (max - min)
		}

		if math.IsNaN(out[id]) || math.IsInf(out[id], 0) {
			out[id] = 0
		}
	}

	return out
}
