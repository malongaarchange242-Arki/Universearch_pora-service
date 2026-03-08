package main

import "math"

// ============================================================
// 📏 NORMALISATION PAR SCORE RAW
// ============================================================

func normalizeByRawScore(nodes []PORANode) {
	if len(nodes) == 0 {
		return
	}

	min := math.MaxFloat64
	max := -math.MaxFloat64

	for _, n := range nodes {
		if n.ScoreRaw < min {
			min = n.ScoreRaw
		}
		if n.ScoreRaw > max {
			max = n.ScoreRaw
		}
	}

	for i := range nodes {
		if min == max {
			nodes[i].Score = 0.5
		} else {
			nodes[i].Score = (nodes[i].ScoreRaw - min) / (max - min)
		}
	}
}

// ============================================================
// 🏆 RANK + PERCENTILE
// ============================================================

func applyRankAndPercentile(nodes []PORANode) {
	total := len(nodes)
	if total == 0 {
		return
	}

	for i := range nodes {
		nodes[i].Rank = i + 1

		if total == 1 {
			nodes[i].Percentile = 100
		} else {
			nodes[i].Percentile = int(
				math.Round(
					(1.0 - float64(i)/float64(total-1)) * 100,
				),
			)
		}
	}
}
