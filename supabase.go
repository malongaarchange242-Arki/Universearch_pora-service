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
		return fmt.Errorf("update universite %s HTTP %d", id, resp.StatusCode())
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
		return fmt.Errorf("update universite %s HTTP %d", id, resp.StatusCode())
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
		return fmt.Errorf("update centre %s HTTP %d", id, resp.StatusCode())
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
