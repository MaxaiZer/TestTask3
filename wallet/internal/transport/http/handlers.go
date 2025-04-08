package http

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"test-task/wallet/internal/domain/models"
	"test-task/wallet/internal/services"
)

type AuthService interface {
	Register(ctx context.Context, username, password, email string) error
	Login(ctx context.Context, username, password string) (string, error)
}

type AuthHandler struct {
	service   AuthService
	validator *validator.Validate
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service, validator: validator.New()}
}

// @Summary Register a new user
// @Description Register a new user with username, password, and email
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body RegisterRequest true "Registration data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /register [post]
func (a *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		slog.Debug("invalid json request", "path", c.Path(), "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload"})
	}

	err := a.validator.Struct(req)
	if err != nil {
		slog.Debug("invalid json request", "path", c.Path(), "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "validation failed: " + err.Error()})
	}

	err = a.service.Register(c.Request().Context(), req.Username, req.Password, req.Email)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, SuccessResponse{Message: "User registered successfully"})
}

// @Summary Login a user
// @Description Login a user and return an authentication token
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Router /login [post]
func (a *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		slog.Debug("invalid json request", "path", c.Path(), "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload"})
	}

	err := a.validator.Struct(req)
	if err != nil {
		slog.Debug("invalid json request", "path", c.Path(), "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "validation failed: " + err.Error()})
	}

	token, err := a.service.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, LoginResponse{Token: token})
}

type WalletService interface {
	GetExchangeRates(ctx context.Context) (map[models.Currency]float64, error)
	GetBalance(ctx context.Context, userID string) (*services.BalanceInfo, error)
	Withdraw(ctx context.Context, userID string, currency models.Currency, amount float64) (*services.BalanceInfo, error)
	Deposit(ctx context.Context, userID string, currency models.Currency, amount float64) (*services.BalanceInfo, error)
	Exchange(ctx context.Context, userID string, from models.Currency, to models.Currency, amount float64) (*services.ExchangeInfo, error)
}

type WalletHandler struct {
	service WalletService
}

func NewWalletHandler(service WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

// @Summary Get the balance of a user
// @Description Get the balance of a user by their token
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BalanceResponse
// @Failure 401 {object} ErrorResponse
// @Router /balance [get]
func (w *WalletHandler) GetBalance(c echo.Context) error {
	userID, err := getUserIdFromToken(c)
	if err != nil {
		return err
	}

	balance, err := w.service.GetBalance(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, BalanceResponse{
		Balance: convertRates(balance.Accounts),
	})
}

// @Summary Deposit money into the user's wallet
// @Description Deposit a specified amount of money into the user's wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param depositRequest body DepositRequest true "Deposit data"
// @Success 200 {object} UpdatedBalanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /wallet/deposit [post]
func (w *WalletHandler) Deposit(c echo.Context) error {
	userID, err := getUserIdFromToken(c)
	if err != nil {
		return err
	}

	var req DepositRequest
	if err = c.Bind(&req); err != nil {
		slog.Debug("invalid json request", "path", c.Path(), "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload"})
	}

	info, err := w.service.Deposit(c.Request().Context(), userID, models.Currency(req.Currency), req.Amount)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, UpdatedBalanceResponse{
		Message:    "Account topped up successfully",
		NewBalance: convertRates(info.Accounts),
	})
}

// @Summary Withdraw money from the user's wallet
// @Description Withdraw a specified amount of money from the user's wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param withdrawRequest body WithdrawRequest true "Withdraw data"
// @Success 200 {object} UpdatedBalanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /wallet/withdraw [post]
func (w *WalletHandler) Withdraw(c echo.Context) error {
	userID, err := getUserIdFromToken(c)
	if err != nil {
		return err
	}

	var req WithdrawRequest
	if err = c.Bind(&req); err != nil {
		slog.Debug("invalid json request", "path", c.Path(), "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload"})
	}

	info, err := w.service.Withdraw(c.Request().Context(), userID, models.Currency(req.Currency), req.Amount)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, UpdatedBalanceResponse{
		Message:    "Withdrawal successful",
		NewBalance: convertRates(info.Accounts),
	})
}

// @Summary Get exchange rates
// @Description Retrieve exchange rates for different currencies
// @Tags wallet
// @Accept json
// @Produce json
// @Success 200 {object} GetRatesResponse
// @Failure 500 {object} ErrorResponse
// @Router /exchange/rates [get]
func (w *WalletHandler) GetRates(c echo.Context) error {
	rates, err := w.service.GetExchangeRates(c.Request().Context())
	if err != nil {
		return err
	}

	res := make(map[string]float64)
	for k, v := range rates {
		res[string(k)] = v
	}

	return c.JSON(http.StatusOK, GetRatesResponse{Rates: res})
}

// @Summary Exchange one currency for another
// @Description Exchange one currency for another in the user's wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param exchangeRequest body ExchangeRequest true "Exchange data"
// @Success 200 {object} ExchangeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /exchange [post]
func (w *WalletHandler) Exchange(c echo.Context) error {

	userID, err := getUserIdFromToken(c)
	if err != nil {
		return err
	}

	var req ExchangeRequest
	if err = c.Bind(&req); err != nil {
		slog.Debug("invalid json request", "path", c.Path(), "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload"})
	}

	info, err := w.service.Exchange(c.Request().Context(), userID, models.Currency(req.FromCurrency),
		models.Currency(req.ToCurrency), req.Amount)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ExchangeResponse{
		Message:         "Exchange successful",
		NewBalance:      convertRates(info.Accounts),
		ExchangedAmount: info.ExchangedAmount,
	})
}

func getUserIdFromToken(c echo.Context) (string, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Token is missing or invalid"})
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to cast claims as jwt.MapClaims")
	}

	extra, ok := claims["extra"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("failed to cast extra")
	}

	id, ok := extra["id"].(string)
	if !ok {
		return "", fmt.Errorf("id field is missing or not a string")
	}
	return id, nil
}

func convertRates(rates map[models.Currency]float64) map[string]float64 {
	formattedRates := make(map[string]float64)
	for k, v := range rates {
		formattedRates[string(k)] = v
	}
	return formattedRates
}
