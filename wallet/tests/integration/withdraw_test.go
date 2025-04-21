package integration

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

func TestWithdraw_Success(t *testing.T) {

	token := getToken(t)

	req1 := myhttp.DepositRequest{
		Amount:   100,
		Currency: "USD",
	}

	_ = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req1, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	req2 := myhttp.WithdrawRequest{
		Amount:   50,
		Currency: "USD",
	}

	resp := mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/withdraw", req2, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, req1.Amount-req2.Amount, resp.NewBalance[req1.Currency])
	assert.Equal(t, "Withdrawal successful", resp.Message)
}

func TestWithdraw_MissingToken(t *testing.T) {

	req := myhttp.WithdrawRequest{
		Amount:   100.5,
		Currency: "USD",
	}

	_ = mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/withdraw", req, http.StatusUnauthorized, nil)
}

func TestWithdraw_NegativeAmount(t *testing.T) {

	token := getToken(t)
	req := myhttp.WithdrawRequest{
		Amount:   -100.5,
		Currency: "USD",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/withdraw", req, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Invalid amount or currency", resp.Error)
}

func TestWithdraw_InvalidCurrency(t *testing.T) {

	token := getToken(t)
	req := myhttp.WithdrawRequest{
		Amount:   100.5,
		Currency: "tugrik",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Invalid amount or currency", resp.Error)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {

	token := getToken(t)

	req1 := myhttp.DepositRequest{
		Amount:   100,
		Currency: "USD",
	}

	_ = mustSend[myhttp.UpdatedBalanceResponse](t, server, "POST",
		apiPrefix+"wallet/deposit", req1, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	req2 := myhttp.WithdrawRequest{
		Amount:   150,
		Currency: "USD",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/withdraw", req2, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Insufficient funds", resp.Error)

	balance := mustSend[myhttp.BalanceResponse](t, server, "GET",
		apiPrefix+"balance", nil, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, req1.Amount, balance.Balance[req1.Currency])
}

func TestWithdraw_NoAccountInThisCurrency(t *testing.T) {

	token := getToken(t)

	req2 := myhttp.WithdrawRequest{
		Amount:   50,
		Currency: "USD",
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"wallet/withdraw", req2, http.StatusBadRequest, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})

	assert.Equal(t, "Insufficient funds", resp.Error)
}
