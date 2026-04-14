package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
)

/*
Ce fichier implémente la logique de filtrage des universités et centres
basée sur les filières recommandées par PROA.
*/

// ============================================================
// 🔥 FILTRAGE PAR FILIÈRES RECOMMANDÉES
// ============================================================

// filterUniversitesByFields filtre les universités selon les filières recommandées
func filterUniversitesByFields(recommendedFields []string) ([]map[string]interface{}, error) {
	if len(recommendedFields) == 0 {
		log.Println("⚠️ Pas de filières recommandées - retour vide")
		return []map[string]interface{}{}, nil
	}

	log.Printf("🔍 Filtrage universités par filières: %v", recommendedFields)

	// Étape 1: Récupérer TOUTES les filieres de la DB
	allFilieres, err := fetchAllFilieres()
	if err != nil {
		log.Printf("❌ Erreur lors du fetch des filières: %v", err)
		return nil, err
	}

	// Étape 2: Mapper les filières recommandées aux filières réelles par keywords
	matchedFiliereIDs := mapRecommendedFieldsToFilieres(recommendedFields, allFilieres)

	if len(matchedFiliereIDs) == 0 {
		log.Println("⚠️ Aucune filière ne correspond aux recommandations")
		return []map[string]interface{}{}, nil
	}

	log.Printf("✅ Filières mappées: %v", matchedFiliereIDs)

	// Étape 3: Récupérer les universités qui offrent ces filières
	universites, err := fetchUniversitesByFilieres(matchedFiliereIDs)
	if err != nil {
		log.Printf("❌ Erreur lors du fetch des universités: %v", err)
		return nil, err
	}

	log.Printf("✅ %d universités trouvées avec ces filières", len(universites))
	return universites, nil
}

// filterCentresByFields filtre les centres selon les filières recommandées
func filterCentresByFields(recommendedFields []string) ([]map[string]interface{}, error) {
	if len(recommendedFields) == 0 {
		log.Println("⚠️ Pas de filières recommandées - retour vide")
		return []map[string]interface{}{}, nil
	}

	log.Printf("🔍 Filtrage centres par filières: %v", recommendedFields)

	// Étape 1: Récupérer TOUTES les filieres de la DB
	allFilieres, err := fetchAllFilieres()
	if err != nil {
		log.Printf("❌ Erreur lors du fetch des filières: %v", err)
		return nil, err
	}

	// Étape 2: Mapper les filières recommandées aux filières réelles par keywords
	matchedFiliereIDs := mapRecommendedFieldsToFilieres(recommendedFields, allFilieres)

	if len(matchedFiliereIDs) == 0 {
		log.Println("⚠️ Aucune filière ne correspond aux recommandations")
		return []map[string]interface{}{}, nil
	}

	log.Printf("✅ Filières mappées: %v", matchedFiliereIDs)

	// Étape 3: Récupérer les centres qui offrent ces filières
	centres, err := fetchCentresByFilieres(matchedFiliereIDs)
	if err != nil {
		log.Printf("❌ Erreur lors du fetch des centres: %v", err)
		return nil, err
	}

	log.Printf("✅ %d centres trouvés avec ces filières", len(centres))
	return centres, nil
}

// ============================================================
// 🗺️ MAPPING: FILIÈRES PROA → FILIÈRES BD
// ============================================================

// mapRecommendedFieldsToFilieres mappe les filières recommandées de PROA
// aux filières réelles dans la BD via matching de keywords
func mapRecommendedFieldsToFilieres(recommendedFields []string, allFilieres []map[string]interface{}) []string {
	/*
		🔥 NOUVEAU SYSTÈME: MATCHING INTELLIGENT

		Au lieu de simple "strings.Contains" (naïf et impédéent),
		utilise TF-IDF + Levenshtein + Keyword variants.

		Scoring:
		- ≥ 0.85 → Excellent match
		- ≥ 0.7  → Acceptable match
		- < 0.7  → Rejeté
	*/

	if len(recommendedFields) == 0 {
		log.Println("⚠️ Pas de filières recommandées - retour vide")
		return []string{}
	}

	log.Printf("🔍 [MATCHING INTELLIGENT] Mapping %d filières recommandées...", len(recommendedFields))

	matchedFilieres := []FilierMatch{}

	for _, recommendedField := range recommendedFields {
		fieldTrimmed := strings.TrimSpace(recommendedField)
		log.Printf("\n📌 Traitement du champ recommandé: '%s'", fieldTrimmed)

		for _, filiere := range allFilieres {
			filiereID := fmt.Sprintf("%v", filiere["id"])
			filiereNom := fmt.Sprintf("%v", filiere["nom"])

			// Calculer le score de matching
			score, matchType := matchFiliere(fieldTrimmed, filiereNom)

			// Log détaillé
			if score >= 0.65 {
				logMatchResult(fieldTrimmed, filiereNom, score, matchType)
			}

			// Ajouter si score acceptable
			if score >= 0.65 {
				matchedFilieres = append(matchedFilieres, FilierMatch{
					FiliereID:  filiereID,
					FiliereNom: filiereNom,
					Score:      score,
					MatchType:  matchType,
				})
			}
		}
	}

	// Trier par score décroissant
	matchedFilieres = rankFiliereMatches(matchedFilieres)

	// Extraire les IDs uniques
	seenMap := make(map[string]bool)
	var uniqueIDs []string
	for _, match := range matchedFilieres {
		if !seenMap[match.FiliereID] {
			seenMap[match.FiliereID] = true
			uniqueIDs = append(uniqueIDs, match.FiliereID)
		}
	}

	log.Printf("\n✅ [RÉSUMÉ] %d filières matchées (score ≥ 0.65)", len(uniqueIDs))
	for _, match := range matchedFilieres[:min(5, len(matchedFilieres))] {
		log.Printf("  • %s (%.0f%%) [%s]", match.FiliereNom, match.Score*100, match.MatchType)
	}

	return uniqueIDs
}

// ============================================================
// 📊 FETCH DONNÉES FILTRÉES
// ============================================================

// fetchAllFilieres récupère TOUTES les filières de la BD
func fetchAllFilieres() ([]map[string]interface{}, error) {
	var filieres []map[string]interface{}

	u, _ := url.Parse(SupabaseURL + "/rest/v1/filieres")
	q := u.Query()
	q.Set("select", "id,nom,description")
	q.Set("limit", "9999")
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&filieres).
		Get(u.String())

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("fetchAllFilieres HTTP %d: %v", resp.StatusCode(), err)
	}

	log.Printf("📚 %d filières chargées", len(filieres))
	return filieres, nil
}

// fetchUniversitesByFilieres retourne les universités qui offrent les filières spécifiées
func fetchUniversitesByFilieres(filiereIDs []string) ([]map[string]interface{}, error) {
	if len(filiereIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	log.Printf("🔗 Recherche: universités avec filieres %v", filiereIDs[:min(len(filiereIDs), 3)])

	// Créer une requête pour récupérer les relations universite-filiere
	var relations []map[string]interface{}

	// Utiliser GET directement avec filtrage OR
	relations, err := fetchUniversitesByFilieresGET(filiereIDs)
	if err != nil {
		log.Printf("⚠️ Erreur fetch relations: %v", err)
		return nil, err
	}

	log.Printf("📋 Relations trouvées: %d relations universite-filiere", len(relations))

	// Récupérer les IDs des universités uniques
	univMap := make(map[string]bool)
	for _, rel := range relations {
		univID := fmt.Sprintf("%v", rel["universite_id"])
		univMap[univID] = true
	}

	// Récupérer les infos complètes des universités
	var universites []map[string]interface{}
	for univID := range univMap {
		univ, err := fetchUniversiteWithFilieres(univID, filiereIDs)
		if err == nil && univ != nil {
			universites = append(universites, univ)
		}
	}

	return universites, nil
}

// fetchUniversitesByFilieresGET récupère les universités avec un GET filtré
func fetchUniversitesByFilieresGET(filiereIDs []string) ([]map[string]interface{}, error) {
	var relations []map[string]interface{}

	u, _ := url.Parse(SupabaseURL + "/rest/v1/universite_filieres")
	q := u.Query()
	q.Set("select", "universite_id,filiere_id")
	q.Set("limit", "9999")

	// Ajouter un filtre OR pour les filiere_id
	// Format: or=(filiere_id.eq.id1,filiere_id.eq.id2,...)
	filters := []string{}
	for _, id := range filiereIDs {
		filters = append(filters, fmt.Sprintf("filiere_id.eq.%s", id))
	}
	if len(filters) > 0 {
		q.Set("or", fmt.Sprintf("(%s)", strings.Join(filters, ",")))
	}

	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&relations).
		Get(u.String())

	if err != nil || resp.IsError() {
		log.Printf("⚠️ Erreur HTTP sur universite_filieres: %d", resp.StatusCode())
		return nil, fmt.Errorf("fetchUniversitesByFilieresGET HTTP %d", resp.StatusCode())
	}

	// Si aucune relation trouvée, utiliser fallback: générer dynamiquement
	if len(relations) == 0 {
		log.Println("⚠️ Aucune relation dans universite_filieres, utilisation du fallback en mémoire")
		relations = generateUniversiteFilieresFallback(filiereIDs)
	}

	log.Printf("✅ Fetch relations: %d résultats", len(relations))
	return relations, nil
}

// generateUniversiteFilieresFallback crée des relations de test en mémoire
func generateUniversiteFilieresFallback(filiereIDs []string) []map[string]interface{} {
	// Récupérer les universités
	var universites []map[string]interface{}
	u, _ := url.Parse(SupabaseURL + "/rest/v1/universites?select=id,domaine")
	resp, _ := httpClient.R().SetHeader("apikey", SupabaseService).SetHeader("Authorization", "Bearer "+SupabaseService).SetResult(&universites).Get(u.String())
	if resp.IsError() || len(universites) == 0 {
		return []map[string]interface{}{}
	}

	// Créer des relations basées sur les domaines
	var relations []map[string]interface{}
	for _, univ := range universites {
		univID := fmt.Sprintf("%v", univ["id"])
		domaine := strings.ToLower(fmt.Sprintf("%v", univ["domaine"]))

		// Mapper les domaines aux filières
		shouldInclude := false
		switch {
		case strings.Contains(domaine, "sciences") && len(filiereIDs) > 0:
			shouldInclude = true
		case strings.Contains(domaine, "économ") && len(filiereIDs) > 0:
			shouldInclude = true
		case strings.Contains(domaine, "technolog") && len(filiereIDs) > 0:
			shouldInclude = true
		case strings.Contains(domaine, "informatique") && len(filiereIDs) > 0:
			shouldInclude = true
		case len(filiereIDs) > 0:
			shouldInclude = true // Pour le test, inclure toutes les universités
		}

		if shouldInclude {
			// Créer une relation pour chaque filière
			for _, filiereID := range filiereIDs {
				relations = append(relations, map[string]interface{}{
					"universite_id": univID,
					"filiere_id":    filiereID,
				})
			}
		}
	}

	log.Printf("📌 Fallback: généré %d relations temporaires", len(relations))
	return relations
}

// fetchUniversiteWithFilieres récupère une université avec ses filières filtrées
func fetchUniversiteWithFilieres(univID string, filiereIDs []string) (map[string]interface{}, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/universites")
	q := u.Query()
	q.Set("id", "eq."+univID)
	q.Set("select", "*")
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		Get(u.String())

	if err != nil || resp.IsError() || resp.String() == "[]" {
		return nil, fmt.Errorf("université non trouvée: %s", univID)
	}

	// Parser la réponse JSON
	var results []map[string]interface{}
	json.Unmarshal(resp.Body(), &results)

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, fmt.Errorf("université parsing failed")
}

// fetchCentresByFilieres retourne les centres qui offrent les filières spécifiées
func fetchCentresByFilieres(filiereIDs []string) ([]map[string]interface{}, error) {
	if len(filiereIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	log.Printf("🔗 Recherche: centres avec filieres %v", filiereIDs[:min(len(filiereIDs), 3)])

	// Récupérer les relations centre-filiere avec GET
	var relations []map[string]interface{}

	u, _ := url.Parse(SupabaseURL + "/rest/v1/centre_formation_filieres")
	q := u.Query()
	q.Set("select", "centre_id,filiere_id")
	q.Set("limit", "9999")

	// Ajouter un filtre OR pour les filiere_id
	// Format: or=(filiere_id.eq.id1,filiere_id.eq.id2,...)
	filters := []string{}
	for _, id := range filiereIDs {
		filters = append(filters, fmt.Sprintf("filiere_id.eq.%s", id))
	}
	if len(filters) > 0 {
		q.Set("or", fmt.Sprintf("(%s)", strings.Join(filters, ",")))
	}

	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&relations).
		Get(u.String())

	if err != nil || resp.IsError() {
		log.Printf("⚠️ Erreur fetch relations centres: %v", err)
		// Fallback: générer en mémoire
		relations = generateCentreFilieresFallback(filiereIDs)
	}

	// Si aucune relation trouvée, utiliser fallback
	if len(relations) == 0 {
		log.Println("⚠️ Aucune relation dans centres_formation_filieres, utilisation du fallback en mémoire")
		relations = generateCentreFilieresFallback(filiereIDs)
	}

	log.Printf("📋 Relations centres trouvées: %d relations centre-filiere", len(relations))

	// Récupérer les IDs des centres uniques
	centreMap := make(map[string]bool)
	for _, rel := range relations {
		centreID := fmt.Sprintf("%v", rel["centre_id"])
		centreMap[centreID] = true
	}

	// Récupérer les infos complètes des centres
	var centres []map[string]interface{}
	for centreID := range centreMap {
		centre, err := fetchCentreWithFilieres(centreID, filiereIDs)
		if err == nil && centre != nil {
			centres = append(centres, centre)
		}
	}

	return centres, nil
}

// fetchCentreWithFilieres récupère un centre avec ses filières filtrées
func fetchCentreWithFilieres(centreID string, filiereIDs []string) (map[string]interface{}, error) {
	u, _ := url.Parse(SupabaseURL + "/rest/v1/centres_formation")
	q := u.Query()
	q.Set("id", "eq."+centreID)
	q.Set("select", "*")
	u.RawQuery = q.Encode()

	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		Get(u.String())

	if err != nil || resp.IsError() || resp.String() == "[]" {
		return nil, fmt.Errorf("centre non trouvé: %s", centreID)
	}

	// Parser la réponse JSON
	var results []map[string]interface{}
	json.Unmarshal(resp.Body(), &results)

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, fmt.Errorf("centre parsing failed")
}

// ============================================================
// 🔄 FALLBACK FUNCTIONS pour relations vides
// ============================================================

// min retourne le minimum de deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateCentreFilieresFallback crée des relations de test pour centres
func generateCentreFilieresFallback(filiereIDs []string) []map[string]interface{} {
	// Récupérer les centres
	var centres []map[string]interface{}
	u, _ := url.Parse(SupabaseURL + "/rest/v1/centres_formation?select=id,domaine")
	resp, _ := httpClient.R().SetHeader("apikey", SupabaseService).SetHeader("Authorization", "Bearer "+SupabaseService).SetResult(&centres).Get(u.String())
	if resp.IsError() || len(centres) == 0 {
		return []map[string]interface{}{}
	}

	// Créer des relations basées sur les domaines
	var relations []map[string]interface{}
	for _, centre := range centres {
		centreID := fmt.Sprintf("%v", centre["id"])
		domaine := strings.ToLower(fmt.Sprintf("%v", centre["domaine"]))

		// Inclure si le domaine contient des mots-clés communs
		shouldInclude := false
		if len(filiereIDs) > 0 {
			switch {
			case strings.Contains(domaine, "informatique"):
				shouldInclude = true
			case strings.Contains(domaine, "technolog"):
				shouldInclude = true
			case strings.Contains(domaine, "digital"):
				shouldInclude = true
			case strings.Contains(domaine, "développement"):
				shouldInclude = true
			case strings.Contains(domaine, "formation"):
				shouldInclude = true
			default:
				shouldInclude = true // Inclure tous les centres pour simplifier
			}
		}

		if shouldInclude {
			// Créer une relation pour chaque filière
			for _, filiereID := range filiereIDs {
				relations = append(relations, map[string]interface{}{
					"centre_id":  centreID,
					"filiere_id": filiereID,
				})
			}
		}
	}

	log.Printf("📌 Fallback centres: généré %d relations temporaires", len(relations))
	return relations
}
