package integration

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	myhttp "test-task/wallet/internal/transport/http"
	"testing"
)

func TestRegister_Success(t *testing.T) {

	req := registerRequestGenerator()

	resp := mustSend[myhttp.SuccessResponse](t, server, "POST",
		apiPrefix+"register", req, http.StatusCreated, nil)

	assert.Equal(t, "User registered successfully", resp.Message)
}

func TestRegister_EmptyName(t *testing.T) {

	req := registerRequestGenerator()
	req.Username = ""

	mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"register", req, http.StatusBadRequest, nil)
}

func TestRegister_EmptyPassword(t *testing.T) {

	req := registerRequestGenerator()
	req.Password = ""

	mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"register", req, http.StatusBadRequest, nil)
}

func TestRegister_EmptyEmail(t *testing.T) {

	req := registerRequestGenerator()
	req.Email = ""

	mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"register", req, http.StatusBadRequest, nil)
}

func TestRegister_InvalidEmail(t *testing.T) {

	req := registerRequestGenerator()
	req.Email = "max"

	mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"register", req, http.StatusBadRequest, nil)
}

func TestRegister_UserAlreadyExists(t *testing.T) {

	req := registerRequestGenerator()

	mustSend[myhttp.SuccessResponse](t, server, "POST",
		apiPrefix+"register", req, http.StatusCreated, nil)

	resp2 := mustSend[myhttp.ErrorResponse](t, server, "POST",
		apiPrefix+"register", req, http.StatusBadRequest, nil)

	assert.Equal(t, "Username or email already exists", resp2.Error)
}
