package main

import (
	"log"
	"strings"
)

/*
INTELLIGENT MATCHING ENGINE FOR PORA

Remplace le simple "strings.Contains" par un système de scoring.

Algorithmes:
1. Levenshtein distance: Pour les typos (ex: "Informatque" vs "Informatique")
2. Keyword matching: Plusieurs variantes (ex: "Data Science" = "data", "science")
3. TF-IDF: Pour les filières multi-mots
4. Scoring: Combinaison des 3 pour une note de 0.0 à 1.0

Score ≥ 0.7 → Match acceptable
Score ≥ 0.85 → Excellent match
*/

// ============================================================
// 1️⃣ LEVENSHTEIN DISTANCE (Tolérance aux typos)
// ============================================================

func levenshteinRatio(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))

	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Calcul de la distance de Levenshtein
	d := make([][]int, len(a)+1)
	for i := range d {
		d[i] = make([]int, len(b)+1)
	}

	for i := 0; i <= len(a); i++ {
		d[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		d[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			d[i][j] = intMin(
				d[i-1][j]+1,
				intMin(d[i][j-1]+1, d[i-1][j-1]+cost),
			)
		}
	}

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	return 1.0 - (float64(d[len(a)][len(b)]) / float64(maxLen))
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================
// 2️⃣ KEYWORD MATCHING (Variantes et synonymes)
// ============================================================

// getKeywordVariants retourne les variantes possibles d'un champ recommandé
func getKeywordVariants(field string) []string {
	fieldLower := strings.ToLower(strings.TrimSpace(field))
	variants := []string{fieldLower}

	// Variantes supplémentaires
	switch fieldLower {
	case "informatique", "génie informatique":
		variants = append(variants, "computer", "science", "software", "dev", "développement")
	case "data science", "sciences de données":
		variants = append(variants, "data", "science", "ia", "machine learning", "analytics")
	case "intelligence artificielle", "ia":
		variants = append(variants, "ia", "ai", "machine", "learning", "intelligence")
	case "réseaux", "réseaux télécoms":
		variants = append(variants, "réseau", "telecom", "communication", "network")
	case "gestion", "management":
		variants = append(variants, "management", "business", "administration")
	case "communication":
		variants = append(variants, "media", "journalisme", "marketing")
	case "médecine", "santé":
		variants = append(variants, "médical", "infirmier", "healthcare")
	case "droit":
		variants = append(variants, "justice", "juridique", "affaires")
	}

	return variants
}

// ============================================================
// 3️⃣ PHRASE MATCHING (Multi-word similarity)
// ============================================================

// tokenize split une phrase en mots importants
func tokenize(phrase string) []string {
	phrase = strings.ToLower(strings.TrimSpace(phrase))

	// Stopwords français à ignorer
	stopwords := map[string]bool{
		"de": true, "du": true, "et": true, "ou": true, "la": true,
		"le": true, "les": true, "un": true, "une": true, "des": true,
		"à": true, "au": true, "&": true,
	}

	words := strings.Fields(phrase)
	var tokens []string
	for _, word := range words {
		if !stopwords[word] && len(word) > 2 {
			tokens = append(tokens, word)
		}
	}
	return tokens
}

// computePhraseScore calcule la similarité entre deux phrases
func computePhraseScore(recommended, actual string) float64 {
	recTokens := tokenize(recommended)
	actTokens := tokenize(actual)

	if len(recTokens) == 0 || len(actTokens) == 0 {
		return 0.0
	}

	// Compter les tokens en commun (Jaccard similarity)
	common := 0
	for _, recToken := range recTokens {
		for _, actToken := range actTokens {
			if recToken == actToken {
				common++
				break
			}
		}
	}

	// Jaccard = intersection / union
	union := len(recTokens) + len(actTokens) - common
	return float64(common) / float64(union)
}

// ============================================================
// 4️⃣ UNIFIED MATCHING SCORE
// ============================================================

// matchFiliere calcule le score d'une filière contre un champ recommandé
// Retourne: (score 0.0-1.0, matchType string)
func matchFiliere(recommendedField string, filiereNom string) (float64, string) {
	recLower := strings.ToLower(strings.TrimSpace(recommendedField))
	filLower := strings.ToLower(strings.TrimSpace(filiereNom))

	// Score parfait: match exact
	if recLower == filLower {
		return 1.0, "EXACT"
	}

	maxScore := 0.0
	matchType := "NONE"

	// 1️⃣ Essayer Levenshtein (tolérance aux typos)
	levScore := levenshteinRatio(recLower, filLower)
	if levScore > 0.85 {
		maxScore = levScore
		matchType = "TYPO"
	}

	// 2️⃣ Essayer keyword matching
	variants := getKeywordVariants(recommendedField)
	for _, variant := range variants {
		if strings.Contains(filLower, variant) || strings.Contains(filLower, variant) {
			// Score basé sur la position et la longueur
			score := 0.8
			if strings.HasPrefix(filLower, variant) {
				score = 0.9 // Bonus si au début
			}
			if score > maxScore {
				maxScore = score
				matchType = "KEYWORD"
			}
		}
	}

	// 3️⃣ Essayer phrase similarity
	phraseScore := computePhraseScore(recommendedField, filiereNom)
	if phraseScore > 0.6 { // Seuil de confiance
		if phraseScore > maxScore {
			maxScore = phraseScore
			matchType = "PHRASE"
		}
	}

	// 0️⃣ Generic field matching fallback
	if isFieldMatch(recommendedField, filiereNom) {
		if maxScore < 0.7 {
			maxScore = 0.7
			matchType = "IS_FIELD"
		}
	}

	// 4️⃣ Fallback: au moins 2 caractères en commun
	if maxScore < 0.5 && countCommonChars(recLower, filLower) > 3 {
		maxScore = 0.6
		matchType = "PARTIAL"
	}

	return maxScore, matchType
}

// countCommonChars compte les caractères en commun
func countCommonChars(a, b string) int {
	common := 0
	aMap := make(map[rune]int)
	for _, r := range a {
		aMap[r]++
	}

	for _, r := range b {
		if aMap[r] > 0 {
			common++
			aMap[r]--
		}
	}
	return common
}

// ============================================================
// 5️⃣ RANKING HELPER
// ============================================================

type FilierMatch struct {
	FiliereID  string
	FiliereNom string
	Score      float64
	MatchType  string
}

// rankFiliereMatches trie les matchs par score décroissant
func rankFiliereMatches(matches []FilierMatch) []FilierMatch {
	// Quick sort simple
	if len(matches) <= 1 {
		return matches
	}

	// Bubble sort pour simplicité (pas critique)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Score > matches[i].Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}

// ============================================================
// 6️⃣ LOGGING & DEBUG
// ============================================================

func logMatchResult(recommendedField string, filiereNom string, score float64, matchType string) {
	symbol := "❌"
	if score >= 0.85 {
		symbol = "✅"
	} else if score >= 0.7 {
		symbol = "⚠️"
	}

	log.Printf("%s [%s] '%s' → '%s' = %.2f", symbol, matchType, recommendedField, filiereNom, score)
}

// Example usage:
/*
func filterUniversitesByFieldsIntelligent(recommendedFields []string) ([]map[string]interface{}, error) {
	allFilieres, _ := fetchAllFilieres()

	matchedMap := make(map[string]FilierMatch) // filiere_id → best match

	for _, recField := range recommendedFields {
		for _, fil := range allFilieres {
			filID := fmt.Sprintf("%v", fil["id"])
			filNom := fmt.Sprintf("%v", fil["nom"])

			score, matchType := matchFiliere(recField, filNom)
			logMatchResult(recField, filNom, score, matchType)

			// Garder le meilleur score pour cette filière
			if score >= 0.7 { // Seuil minimum
				if existing, found := matchedMap[filID]; !found || score > existing.Score {
					matchedMap[filID] = FilierMatch{
						FiliereID: filID,
						FiliereNom: filNom,
						Score: score,
						MatchType: matchType,
					}
				}
			}
		}
	}

	// Convertir en slice et trier
	var matches []FilierMatch
	for _, m := range matchedMap {
		matches = append(matches, m)
	}
	matches = rankFiliereMatches(matches)

	// Retourner les universités qui offrent ces filières
	...
}
*/
