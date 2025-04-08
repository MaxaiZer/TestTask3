package integration

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"test-task/wallet/internal/domain/models"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

func TestExchange_Success(t *testing.T) {

	token := getToken(t)

	req1 := myhttp.DepositRequest{
		Amount:   100,
		Currency: "USD",
	}

	_ = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req1, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	req2 := myhttp.DepositRequest{
		Amount:   100,
		Currency: "EUR",
	}

	_ = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req2, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	req3 := myhttp.ExchangeRequest{
		Amount:       50,
		FromCurrency: "USD",
		ToCurrency:   "EUR",
	}

	resp := mustSend[myhttp.ExchangeResponse](t, server, "POST",
		apiPrefix+"exchange", req3, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	exchangedAmount := req3.Amount * rates[models.Currency(req2.Currency)]
	assert.Equal(t, req1.Amount-req3.Amount, resp.NewBalance[req1.Currency])
	assert.Equal(t, req2.Amount+exchangedAmount, resp.NewBalance[req2.Currency])
	assert.Equal(t, exchangedAmount, resp.ExchangedAmount)
}

func TestExchange_NegativeAmount(t *testing.T) {

	token := getToken(t)

	req1 := myhttp.DepositRequest{
		Amount:   100,
		Currency: "USD",
	}

	_ = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req1, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	req2 := myhttp.DepositRequest{
		Amount:   100,
		Currency: "EUR",
	}

	_ = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req2, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	req3 := myhttp.ExchangeRequest{
		Amount:       -50,
		FromCurrency: "USD",
		ToCurrency:   "EUR",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"exchange", req3, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Invalid amount or currency", resp.Error)
}

func TestExchange_SameCurrencies(t *testing.T) {

	token := getToken(t)

	req1 := myhttp.DepositRequest{
		Amount:   100,
		Currency: "USD",
	}

	_ = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req1, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	req2 := myhttp.ExchangeRequest{
		Amount:       50,
		FromCurrency: "USD",
		ToCurrency:   "USD",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"exchange", req2, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Invalid amount or currency", resp.Error)
}
