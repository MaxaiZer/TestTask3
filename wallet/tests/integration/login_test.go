package integration

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

func TestLogin_Success(t *testing.T) {

	registerReq := registerRequestGenerator()

	_ = mustSend[myhttp.SuccessResponse](t, server, "POST",
		apiPrefix+"register", registerReq, http.StatusCreated, nil)

	loginReq := myhttp.LoginRequest{
		Username: registerReq.Username,
		Password: registerReq.Password,
	}

	resp := mustSend[myhttp.LoginResponse](t, server, "POST",
		apiPrefix+"login", loginReq, http.StatusOK, nil)
	assert.NotEmpty(t, resp.Token)
}

func TestLogin_UserDoesntExist(t *testing.T) {

	registerReq := registerRequestGenerator()

	loginReq := myhttp.LoginRequest{
		Username: registerReq.Username,
		Password: registerReq.Password,
	}

	resp := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"login", loginReq, http.StatusUnauthorized, nil)
	assert.Equal(t, "Invalid username or password", resp.Error)
}

func TestLogin_EmptyName(t *testing.T) {

	registerReq := registerRequestGenerator()

	_ = mustSend[myhttp.SuccessResponse](t, server, "POST",
		apiPrefix+"register", registerReq, http.StatusCreated, nil)

	loginReq := myhttp.LoginRequest{
		Password: registerReq.Password,
	}

	_ = mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"login", loginReq, http.StatusBadRequest, nil)
}

func TestLogin_EmptyPassword(t *testing.T) {
	registerReq := registerRequestGenerator()

	_ = mustSend[myhttp.SuccessResponse](t, server, "POST",
		apiPrefix+"register", registerReq, http.StatusCreated, nil)

	loginReq := myhttp.LoginRequest{
		Password: registerReq.Password,
	}

	_ = mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"login", loginReq, http.StatusBadRequest, nil)
}
