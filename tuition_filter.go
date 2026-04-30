package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

type TuitionFee struct {
	UniversiteID string  `json:"universite_id"`
	Level        string  `json:"level"`
	Pole         string  `json:"pole"`
	MonthlyPrice float64 `json:"monthly_price"`
	YearlyPrice  float64 `json:"yearly_price"`
	Currency     string  `json:"currency"`
}

func filterUniversitesByBudget(items []map[string]interface{}, maxMonthlyPrice float64) ([]map[string]interface{}, error) {
	if len(items) == 0 {
		return []map[string]interface{}{}, nil
	}
	if maxMonthlyPrice <= 0 {
		return items, enrichUniversitesWithFeeInfo(items)
	}

	ids := extractUniversiteIDs(items)
	feesByUniversite, err := fetchTuitionFeesByUniversites(ids)
	if err != nil {
		return nil, err
	}

	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		universiteID, _ := item["id"].(string)
		minFee, ok := minTuitionFee(feesByUniversite[universiteID])
		if !ok {
			log.Printf("💰 Université %s exclue: aucun frais_scolarite renseigné", universiteID)
			continue
		}

		addFeeInfo(item, minFee)
		if effectiveMonthlyPrice(minFee) <= maxMonthlyPrice {
			filtered = append(filtered, item)
			continue
		}

		log.Printf("💰 Université %s exclue: frais mensuel %.0f > budget %.0f", universiteID, effectiveMonthlyPrice(minFee), maxMonthlyPrice)
	}

	return filtered, nil
}

func enrichUniversitesWithFeeInfo(items []map[string]interface{}) error {
	if len(items) == 0 {
		return nil
	}

	ids := extractUniversiteIDs(items)
	feesByUniversite, err := fetchTuitionFeesByUniversites(ids)
	if err != nil {
		return err
	}

	for _, item := range items {
		universiteID, _ := item["id"].(string)
		if minFee, ok := minTuitionFee(feesByUniversite[universiteID]); ok {
			addFeeInfo(item, minFee)
		}
	}

	return nil
}

func fetchTuitionFeesByUniversites(universiteIDs []string) (map[string][]TuitionFee, error) {
	out := make(map[string][]TuitionFee)
	universiteIDs = uniqueOrderedIDs(universiteIDs)
	if len(universiteIDs) == 0 {
		return out, nil
	}

	u, _ := url.Parse(SupabaseURL + "/rest/v1/frais_scolarite")
	q := u.Query()
	q.Set("select", "universite_id,level,pole,monthly_price,yearly_price,currency")
	q.Set("universite_id", fmt.Sprintf("in.(%s)", strings.Join(universiteIDs, ",")))
	q.Set("limit", "9999")
	u.RawQuery = q.Encode()

	var rows []TuitionFee
	resp, err := httpClient.R().
		SetHeader("apikey", SupabaseService).
		SetHeader("Authorization", "Bearer "+SupabaseService).
		SetResult(&rows).
		Get(u.String())

	if err != nil || resp.IsError() {
		status := 0
		if resp != nil {
			status = resp.StatusCode()
		}
		return nil, fmt.Errorf("fetch frais_scolarite HTTP %d: %v", status, err)
	}

	for _, row := range rows {
		if row.UniversiteID == "" || effectiveMonthlyPrice(row) <= 0 {
			continue
		}
		if row.Currency == "" {
			row.Currency = "XAF"
		}
		out[row.UniversiteID] = append(out[row.UniversiteID], row)
	}

	return out, nil
}

func extractUniversiteIDs(items []map[string]interface{}) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if id, ok := item["id"].(string); ok && strings.TrimSpace(id) != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func minTuitionFee(fees []TuitionFee) (TuitionFee, bool) {
	if len(fees) == 0 {
		return TuitionFee{}, false
	}

	minFee := fees[0]
	for _, fee := range fees[1:] {
		if effectiveMonthlyPrice(fee) > 0 && effectiveMonthlyPrice(fee) < effectiveMonthlyPrice(minFee) {
			minFee = fee
		}
	}

	return minFee, true
}

func effectiveMonthlyPrice(fee TuitionFee) float64 {
	if fee.MonthlyPrice > 0 {
		return fee.MonthlyPrice
	}
	if fee.YearlyPrice > 0 {
		return fee.YearlyPrice / 12
	}
	return 0
}

func addFeeInfo(item map[string]interface{}, fee TuitionFee) {
	item["min_yearly_price"] = fee.YearlyPrice
	item["min_monthly_price"] = effectiveMonthlyPrice(fee)
	item["fee_level"] = fee.Level
	item["fee_pole"] = fee.Pole
	item["fee_currency"] = fee.Currency
}
