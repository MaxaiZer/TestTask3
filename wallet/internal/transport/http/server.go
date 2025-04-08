package http

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"log/slog"
	"net/http"
	errs "test-task/wallet/internal/domain/errors"
	"test-task/wallet/internal/tracing"
)

type Config struct {
	ServiceName   string
	JwtSecret     string
	LaunchSwagger bool
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// @title Wallet API
// @version 1.0
// @description This is the API for managing wallets and performing actions like deposits, withdrawals, and exchange rates.
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func NewServer(config Config, walletService WalletService, authService AuthService) *echo.Echo {

	e := echo.New()
	jwtMiddleware := echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(config.JwtSecret),
		SigningMethod: "HS512",
	})

	v := validator.New()
	e.Validator = &customValidator{validator: v}

	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware(config.ServiceName))

	e.HTTPErrorHandler = errorHandler

	auth := NewAuthHandler(authService)
	wallet := NewWalletHandler(walletService)

	api := e.Group("/api/v1")

	api.POST("/register", auth.Register)
	api.POST("/login", auth.Login)

	api.GET("/exchange/rates", wallet.GetRates)
	api.POST("/exchange", wallet.Exchange, jwtMiddleware)
	api.POST("/wallet/withdraw", wallet.Withdraw, jwtMiddleware)
	api.POST("/wallet/deposit", wallet.Deposit, jwtMiddleware)
	api.GET("/balance", wallet.GetBalance, jwtMiddleware)

	if config.LaunchSwagger {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	return e
}

func errorHandler(err error, c echo.Context) {

	if c.Response().Committed {
		return
	}

	traceInfo, traceErr := tracing.GetTraceInfo(c.Request().Context())
	traceID := "unknown_trace_id"
	if traceErr != nil {
		slog.Warn("failed to retrieve traceID", "error", traceErr)
	} else {
		traceID = traceInfo.TraceID
	}

	code := http.StatusInternalServerError
	message := "Internal Server Error"

	switch {
	case errors.Is(err, errs.UserAlreadyExists):
		code = http.StatusBadRequest
		message = "Username or email already exists"
	case errors.Is(err, errs.UserNotExists) || errors.Is(err, errs.WrongPassword):
		code = http.StatusUnauthorized
		message = "Invalid username or password"
	case errors.Is(err, errs.InvalidAmount) || errors.Is(err, errs.InvalidCurrency):
		code = http.StatusBadRequest
		message = "Invalid amount or currency"
	case errors.Is(err, errs.InsufficientFunds):
		code = http.StatusBadRequest
		message = "Insufficient funds"
	case errors.Is(err, echojwt.ErrJWTInvalid):
		code = http.StatusUnauthorized
		message = "Invalid JWT"
		slog.Debug("invalid jwt", "path", c.Path(), "error", err)
	case errors.Is(err, echojwt.ErrJWTMissing):
		code = http.StatusUnauthorized
		message = "Missing JWT"
	case errors.Is(err, echo.ErrNotFound):
		code = http.StatusNotFound
		message = "Not Found"
	}

	if code == http.StatusInternalServerError {
		slog.Error("error occurred", "traceId", traceID, "path", c.Path(), "error", err)
	}
	c.JSON(code, ErrorResponse{Error: message})
}
