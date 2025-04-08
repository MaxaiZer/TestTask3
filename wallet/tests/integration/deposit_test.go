package integration

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

func TestDeposit_Success(t *testing.T) {

	token := getToken(t)
	req := myhttp.DepositRequest{
		Amount:   100.5,
		Currency: "USD",
	}

	resp := mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, req.Amount, resp.NewBalance[req.Currency])
	assert.Equal(t, "Account topped up successfully", resp.Message)

	resp = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, req.Amount*2, resp.NewBalance[req.Currency])
}

func TestDeposit_MissingToken(t *testing.T) {

	req := myhttp.DepositRequest{
		Amount:   100.5,
		Currency: "USD",
	}

	_ = mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req, http.StatusUnauthorized, nil)
}

func TestDeposit_NegativeAmount(t *testing.T) {

	token := getToken(t)
	req := myhttp.DepositRequest{
		Amount:   -100.5,
		Currency: "USD",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Invalid amount or currency", resp.Error)
}

func TestDeposit_InvalidCurrency(t *testing.T) {

	token := getToken(t)
	req := myhttp.DepositRequest{
		Amount:   100.5,
		Currency: "tugrik",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Invalid amount or currency", resp.Error)
}

func TestDeposit_Concurrently(t *testing.T) {

	token := getToken(t)
	req := myhttp.DepositRequest{
		Amount:   100,
		Currency: "USD",
	}
	requests := 100
	expectedSum := req.Amount * float64(requests)

	var returnedBalances []float64
	var mu sync.Mutex

	wg := sync.WaitGroup{}
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
				apiPrefix+"wallet/deposit", req, http.StatusOK, func(request *http.Request) {
					request.Header.Set("Authorization", "Bearer "+token)
				})
			mu.Lock()
			returnedBalances = append(returnedBalances, resp.NewBalance[req.Currency])
			mu.Unlock()
		}()
	}

	wg.Wait()

	for i := 1; i < requests; i++ {
		assert.Equal(t, returnedBalances[i-1]+req.Amount, returnedBalances[i])
	}

	balance := mustSend[myhttp.BalanceResponse](t, server, "GET",
		apiPrefix+"balance", nil, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, expectedSum, balance.Balance[req.Currency])
}
