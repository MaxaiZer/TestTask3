package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"test-task/wallet/internal/domain/models"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

var unexpectedCode = errors.New("unexpected code")
var parsingError = errors.New("could not parse response")

func mustSend[Resp any](t require.TestingT, handler http.Handler, method, url string, request any, expectedCode int,
	transform func(*http.Request)) *Resp {

	if transform == nil {
		transform = func(*http.Request) {}
	}

	body, _ := json.Marshal(request)

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	transform(req)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != expectedCode {
		return returnResponseWithCheck[Resp](t, nil, fmt.Errorf("%w: %d", unexpectedCode, w.Code))
	}

	var response Resp
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		return returnResponseWithCheck[Resp](t, nil, fmt.Errorf("%w: %s", parsingError, err))
	}
	return returnResponseWithCheck[Resp](t, &response, nil)
}

func returnResponseWithCheck[Res any](t require.TestingT, resp *Res, err error) *Res {
	require.NoError(t, err)
	require.NotNil(t, resp)
	return resp
}

func convertRates(rates map[models.Currency]float64) map[string]float64 {
	res := map[string]float64{}
	for k, v := range rates {
		res[string(k)] = v
	}
	return res
}

var registerRequestGenerator = createRegisterRequestGenerator()

func createRegisterRequestGenerator() func() myhttp.RegisterRequest {
	var number atomic.Int64

	return func() myhttp.RegisterRequest {

		value := number.Add(1)

		request := myhttp.RegisterRequest{
			Username: "max" + strconv.Itoa(int(value)),
			Password: "1234",
			Email:    "max" + strconv.Itoa(int(value)) + "@mail.ru",
		}
		return request
	}
}

func getToken(t *testing.T) string {
	registerReq := registerRequestGenerator()

	_ = mustSend[myhttp.SuccessResponse](t, server, "POST",
		apiPrefix+"register", registerReq, http.StatusCreated, nil)

	loginReq := myhttp.LoginRequest{
		Username: registerReq.Username,
		Password: registerReq.Password,
	}

	resp := mustSend[myhttp.LoginResponse](t, server, "POST",
		apiPrefix+"login", loginReq, http.StatusOK, nil)
	return resp.Token
}
