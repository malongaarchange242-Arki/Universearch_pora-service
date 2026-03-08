package main

import "math"

//
// ============================================================
// 🔁 PORA TREND (Δ SCORE)
// ============================================================
//

type PORATrend struct {
	Current  float64 `json:"current"`
	Previous float64 `json:"previous"`
	Delta    float64 `json:"delta"`
	Status   string  `json:"status"` // up | down | stable
}

//
// ------------------------------------------------------------
// 📈 CALCUL DU TREND
// ------------------------------------------------------------
//

func computeTrend(current, previous float64) PORATrend {
	delta := current - previous

	status := "stable"
	if delta > 0.01 {
		status = "up"
	} else if delta < -0.01 {
		status = "down"
	}

	return PORATrend{
		Current:  round(current, 4),
		Previous: round(previous, 4),
		Delta:    round(delta, 4),
		Status:   status,
	}
}

//
// ------------------------------------------------------------
// 🔢 HELPER ROUND
// ------------------------------------------------------------
//

func round(v float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(v*p) / p
}
