package main

import (
	"math"
	"sort"
)

//
// ============================================================
// 🧠 PORA — NODE GLOBAL (UNIVERSITÉ + CENTRE)
// ============================================================
//
// RÈGLES :
// - ScoreRaw  → score PORA réel (base trend & historique)
// - Score     → score normalisé (PAR TYPE)
// - Trend     → calculé AVANT normalisation
//

//
// ============================================================
// 🔗 ASSEMBLAGE DES NŒUDS GLOBAUX
// ============================================================
//

func fetchGlobalNodes() ([]PORANode, error) {
	nodes := []PORANode{}

	// ---------- UNIVERSITÉS ----------
	unis, err := fetchUniversites()
	if err != nil {
		return nil, err
	}

	for _, u := range unis {
		nodes = append(nodes, PORANode{
			ID:       u.ID,
			Type:     "universite",
			ScoreRaw: u.ScorePora,
			Score:    u.ScorePora,
			Trend:    computeTrend(u.ScorePora, u.ScorePoraPrev),
			Detail:   u.ScoreDetails,
		})
	}

	// ---------- CENTRES ----------
	centres, err := fetchCentresFormation()
	if err != nil {
		return nil, err
	}

	for _, c := range centres {
		nodes = append(nodes, PORANode{
			ID:       c.ID,
			Type:     "centre",
			ScoreRaw: c.ScorePora,
			Score:    c.ScorePora,
			Trend:    computeTrend(c.ScorePora, c.ScorePoraPrev),
			Detail:   c.ScoreDetails,
		})
	}

	return nodes, nil
}

//
// ============================================================
// 📏 NORMALISATION + RANK PAR TYPE
// ============================================================
//

func normalizeAndRankByType(nodes []PORANode) []PORANode {
	if len(nodes) == 0 {
		return nodes
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

	// Normalisation
	for i := range nodes {
		if min == max {
			nodes[i].Score = 0.5
		} else {
			nodes[i].Score = (nodes[i].ScoreRaw - min) / (max - min)
		}
	}

	// Tri décroissant
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Score > nodes[j].Score
	})

	// Rank + Percentile
	total := len(nodes)
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

	return nodes
}

//
// ============================================================
// 🌍 RANKING GLOBAL PUBLIC (PAR TYPE)
// ============================================================
//

func RankGlobal() ([]PORANode, error) {
	nodes, err := fetchGlobalNodes()
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nodes, nil
	}

	var universites []PORANode
	var centres []PORANode

	for _, n := range nodes {
		if n.Type == "universite" {
			universites = append(universites, n)
		} else {
			centres = append(centres, n)
		}
	}

	// 🔥 Ranking PAR TYPE
	universites = normalizeAndRankByType(universites)
	centres = normalizeAndRankByType(centres)

	// Fusion finale
	return append(universites, centres...), nil
}
