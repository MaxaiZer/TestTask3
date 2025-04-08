package integration

import "test-task/exchanger/internal/models"

func convertRates(rates map[models.Currency]float64) map[string]float64 {
	res := map[string]float64{}
	for k, v := range rates {
		res[string(k)] = v
	}
	return res
}
