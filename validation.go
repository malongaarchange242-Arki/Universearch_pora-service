/*
IMPLÉMENTATION PHASE 1: VALIDATION & WEIGHTED MATCHING (PORA)
==============================================================

Fichier: services/pora-service/validation.go
Module de validation de données universités
*/

package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// ============================================================
// 1️⃣ STRUCTURES POUR VALIDATION
// ============================================================

type ValidationError struct {
	Field    string
	Message  string
	Severity string // "ERROR", "WARNING", "INFO"
}

type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
	Score    float64 // 0-1: confiance dans les données
}

// ============================================================
// 2️⃣ VALIDATEURS SPÉCIFIQUES
// ============================================================

func validateUniversityName(name string) ValidationError {
	if name == "" {
		return ValidationError{
			Field:    "name",
			Message:  "Name cannot be empty",
			Severity: "ERROR",
		}
	}
	if len(name) < 3 {
		return ValidationError{
			Field:    "name",
			Message:  fmt.Sprintf("Name too short: %s", name),
			Severity: "WARNING",
		}
	}
	return ValidationError{}
}

func validateInstitutionID(id string) ValidationError {
	if id == "" {
		return ValidationError{
			Field:    "id",
			Message:  "ID cannot be empty",
			Severity: "ERROR",
		}
	}
	// UUID format check (basic)
	uuidRegex := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	if !uuidRegex.MatchString(id) {
		log.Printf("⚠️  ID not in UUID format: %s", id)
		return ValidationError{} // Warning, not error
	}
	return ValidationError{}
}

func validateFilieres(filieres []string, filiereCatalog map[string]bool) []ValidationError {
	var errors []ValidationError

	if len(filieres) == 0 {
		errors = append(errors, ValidationError{
			Field:    "filieres",
			Message:  "At least one filiere required",
			Severity: "WARNING",
		})
		return errors
	}

	for _, filiere := range filieres {
		filiere = strings.TrimSpace(filiere)

		if filiere == "" {
			errors = append(errors, ValidationError{
				Field:    "filieres",
				Message:  "Empty filiere in list",
				Severity: "WARNING",
			})
			continue
		}

		// Check if filiere exists in catalog
		if _, exists := filiereCatalog[strings.ToLower(filiere)]; !exists {
			log.Printf("⚠️  Filiere not in catalog: %s", filiere)
			errors = append(errors, ValidationError{
				Field:    "filieres",
				Message:  fmt.Sprintf("Unknown filiere: %s", filiere),
				Severity: "WARNING",
			})
		}
	}

	return errors
}

func validateScore(score float64) ValidationError {
	if score < 0 || score > 1 {
		return ValidationError{
			Field:    "score",
			Message:  fmt.Sprintf("Score must be 0-1, got %f", score),
			Severity: "ERROR",
		}
	}
	return ValidationError{}
}

func validateLocation(location string) ValidationError {
	if location == "" {
		return ValidationError{
			Field:    "location",
			Message:  "Location cannot be empty",
			Severity: "WARNING",
		}
	}
	return ValidationError{}
}

func validateCapacity(capacity int) ValidationError {
	if capacity <= 0 {
		return ValidationError{
			Field:    "capacity",
			Message:  fmt.Sprintf("Capacity must be > 0, got %d", capacity),
			Severity: "WARNING",
		}
	}
	if capacity < 10 {
		return ValidationError{
			Field:    "capacity",
			Message:  fmt.Sprintf("Very small capacity: %d", capacity),
			Severity: "INFO",
		}
	}
	return ValidationError{}
}

// ============================================================
// 3️⃣ VALIDATION GLOBALE
// ============================================================

func validateUniversityData(uni map[string]interface{}, filiereCatalog map[string]bool) ValidationResult {
	var errors []ValidationError
	var warnings []ValidationError

	// Extract fields from map
	id, _ := uni["id"].(string)
	nom, _ := uni["nom"].(string)
	score, _ := uni["score_pora"].(float64)

	log.Printf("🔍 Validating university: %s", nom)

	// 1. Validate ID
	if err := validateInstitutionID(id); err.Severity != "" {
		if err.Severity == "ERROR" {
			errors = append(errors, err)
		} else {
			warnings = append(warnings, err)
		}
	}

	// 2. Validate Name
	if err := validateUniversityName(nom); err.Severity != "" {
		if err.Severity == "ERROR" {
			errors = append(errors, err)
		} else {
			warnings = append(warnings, err)
		}
	}

	// 3. Validate Score
	if err := validateScore(score); err.Severity != "" {
		errors = append(errors, err)
	}

	// Note: Filieres, Location, Capacity validation removed as they're not in the current data structure
	// They can be added back when the data structure is enriched

	// Calculate confidence score
	confidence := 1.0
	if len(errors) > 0 {
		confidence -= float64(len(errors)) * 0.2
	}
	if len(warnings) > 0 {
		confidence -= float64(len(warnings)) * 0.05
	}
	if confidence < 0 {
		confidence = 0
	}

	valid := len(errors) == 0

	if valid {
		log.Printf("✅ Validation passed (confidence: %.0f%%)", confidence*100)
	} else {
		log.Printf("❌ Validation failed: %d errors, %d warnings", len(errors), len(warnings))
		for _, e := range errors {
			log.Printf("   ERROR: %s - %s", e.Field, e.Message)
		}
	}

	return ValidationResult{
		Valid:    valid,
		Errors:   errors,
		Warnings: warnings,
		Score:    confidence,
	}
}

// ============================================================
// 4️⃣ FILTERED RECOMMENDATIONS (avec validation)
// ============================================================

func getRecommendationsWithValidation(
	recommendations []map[string]interface{},
	filiereCatalog map[string]bool,
) ([]map[string]interface{}, []ValidationError) {

	var validatedRecommendations []map[string]interface{}
	var allErrors []ValidationError

	log.Printf("🔍 Validating %d recommendations", len(recommendations))

	for i, uni := range recommendations {
		result := validateUniversityData(uni, filiereCatalog)

		nom, _ := uni["nom"].(string)

		if result.Valid {
			validatedRecommendations = append(validatedRecommendations, uni)
			log.Printf("   [%d/%d] ✅ %s", i+1, len(recommendations), nom)
		} else {
			log.Printf("   [%d/%d] ❌ %s (skipped - %d errors)",
				i+1, len(recommendations), nom, len(result.Errors))
			allErrors = append(allErrors, result.Errors...)
		}

		if len(result.Warnings) > 0 {
			log.Printf("        ⚠️  %d warnings", len(result.Warnings))
		}
	}

	log.Printf("📊 Result: %d/%d recommendations passed validation",
		len(validatedRecommendations), len(recommendations))

	return validatedRecommendations, allErrors
}

// ============================================================
// 5️⃣ WEIGHTED MATCHING ENGINE
// ============================================================

type MatchingWeights struct {
	DomainMatch      float64 // Poids pour match du domaine principal
	SecondaryMatch   float64 // Poids pour match du domaine secondaire
	TertiaryMatch    float64 // Poids pour match du 3e domaine
	LocationBonus    float64 // Bonus localisation
	ReputationFactor float64 // Facteur réputation
}

// GetDefaultWeights retourne les poids par défaut
func GetDefaultWeights() MatchingWeights {
	return MatchingWeights{
		DomainMatch:      0.50, // Top domaine = 50%
		SecondaryMatch:   0.30, // 2e domaine = 30%
		TertiaryMatch:    0.20, // 3e domaine = 20%
		LocationBonus:    0.10, // +10% si même région
		ReputationFactor: 0.05, // Petit ajustement réputation
	}
}

func calculateWeightedScore(
	uni map[string]interface{},
	domainScores map[string]float64,
	weights MatchingWeights,
	userLocation string,
) float64 {

	score := 0.0

	// For now, we don't have filieres in the map structure
	// This is a placeholder for when the data structure is enriched
	// The weighted matching is already implemented in handlers.go

	// Extract current score if available
	if currentScore, ok := uni["score_pora"].(float64); ok {
		score = currentScore
	}

	// Clamp à [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// ============================================================
// 6️⃣ EXEMPLES D'UTILISATION
// ============================================================

/*
// Dans le handler /recommendations:

func getRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	profile := extractProfileFromRequest(r)  // User profile
	userLocation := profile.Location

	// 1. Get initial recommendations (existing logic)
	recommendations := matchUniversities(profile)

	// 2. Validate each recommendation
	filiereCatalog := getFiliereCatalog()  // Cache it!
	validated, errors := getRecommendationsWithValidation(
		recommendations,
		filiereCatalog,
	)

	// 3. Apply weighted matching (already implemented in handlers.go)
	weights := GetDefaultWeights()
	for i := range validated {
		validated[i]["score_pora"] = calculateWeightedScore(
			validated[i],
			profile.DomainScores,
			weights,
			userLocation,
		)
	}

	// 4. Sort by new scores
	sort.Slice(validated, func(i, j int) bool {
		scoreI, _ := validated[i]["score_pora"].(float64)
		scoreJ, _ := validated[j]["score_pora"].(float64)
		return scoreI > scoreJ
	})

	// 5. Return top 5 with validation metadata
	response := map[string]interface{}{
		"recommendations": validated[:min(5, len(validated))],
		"validation_summary": map[string]interface{}{
			"total_checked": len(recommendations),
			"passed":        len(validated),
			"errors":        len(errors),
		},
	}

	json.NewEncoder(w).Encode(response)
}
*/
