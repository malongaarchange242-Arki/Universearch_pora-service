package main

import "sort"

//
// ============================================================
// 🏫 RANKING UNIVERSITÉS
// ============================================================
//

func RankUniversites() ([]PORANode, error) {
	nodes, err := fetchGlobalNodes()
	if err != nil {
		return nil, err
	}

	// 🔹 Filtrage universités
	unis := []PORANode{}
	for _, n := range nodes {
		if n.Type == "universite" {
			unis = append(unis, n)
		}
	}

	if len(unis) == 0 {
		return unis, nil
	}

	// 1️⃣ Normalisation spécifique universités (SUR SCORE BRUT)
	normalizeByRawScore(unis)

	// 2️⃣ Tri
	sort.Slice(unis, func(i, j int) bool {
		return unis[i].Score > unis[j].Score
	})

	// 3️⃣ Rank + Percentile
	applyRankAndPercentile(unis)

	return unis, nil
}

//
// ============================================================
// 🏢 RANKING CENTRES DE FORMATION
// ============================================================
//

func RankCentresFormation() ([]PORANode, error) {
	nodes, err := fetchGlobalNodes()
	if err != nil {
		return nil, err
	}

	// 🔹 Filtrage centres
	centres := []PORANode{}
	for _, n := range nodes {
		if n.Type == "centre" {
			centres = append(centres, n)
		}
	}

	if len(centres) == 0 {
		return centres, nil
	}

	// 1️⃣ Normalisation spécifique centres (SUR SCORE BRUT)
	normalizeByRawScore(centres)

	// 2️⃣ Tri
	sort.Slice(centres, func(i, j int) bool {
		return centres[i].Score > centres[j].Score
	})

	// 3️⃣ Rank + Percentile
	applyRankAndPercentile(centres)

	return centres, nil
}
