package integration

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

func TestGetRates(t *testing.T) {

	resp := mustSend[myhttp.GetRatesResponse](t, server, "GET", apiPrefix+"exchange/rates", nil,
		http.StatusOK, nil)
	assert.Equal(t, resp.Rates, convertRates(rates))
}
