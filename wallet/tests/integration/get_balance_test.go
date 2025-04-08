package integration

import (
	"net/http"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

func TestGetBalance_Success(t *testing.T) {

	token := getToken(t)

	_ = mustSend[myhttp.BalanceResponse](t, server, "GET",
		apiPrefix+"balance", nil, http.StatusOK, func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
		})
}

func TestGetBalance_MissingToken(t *testing.T) {

	_ = mustSend[myhttp.BalanceResponse](t, server, "GET",
		apiPrefix+"balance", nil, http.StatusUnauthorized, nil)
}
