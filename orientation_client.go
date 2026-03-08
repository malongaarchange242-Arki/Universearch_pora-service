package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ==================================================
// ORIENTATION — COMPUTE FULL PROFILE (PROA)
// ==================================================

// Réponse complète du microservice Orientation
type OrientationComputeResponse struct {
	UserID     string    `json:"user_id"`
	Profile    []float64 `json:"profile"`
	Confidence float64   `json:"confidence"`
}

// Appel utilisé par la page ORIENTATION (quiz utilisateur)
func CallOrientationCompute(
	userID string,
	quizVersion string,
	responses map[string]float64,
) (*OrientationComputeResponse, error) {

	if OrientationServiceURL == "" {
		return nil, fmt.Errorf("ORIENTATION_SERVICE_URL not set")
	}

	payload := map[string]any{
		"user_id":      userID,
		"quiz_version": quizVersion,
		"responses":    responses,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		OrientationServiceURL+"/orientation/compute",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf(
			"orientation compute error %d → %s",
			resp.StatusCode,
			buf.String(),
		)
	}

	var result OrientationComputeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode orientation compute response: %w", err)
	}

	return &result, nil
}
