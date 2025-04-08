package http

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type DepositRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0" example:"15"`
	Currency string  `json:"currency" validate:"required" example:"USD"`
}

type WithdrawRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0" example:"15"`
	Currency string  `json:"currency" validate:"required" example:"USD"`
}

type BalanceResponse struct {
	Balance map[string]float64 `json:"balance" example:"USD:20.0,EUR:1.5,RUB:15.0"`
}

type UpdatedBalanceResponse struct {
	Message    string             `json:"message"`
	NewBalance map[string]float64 `json:"new_balance" example:"USD:20.0,EUR:1.5,RUB:15.0"`
}

type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" validate:"required" example:"USD"`
	ToCurrency   string  `json:"to_currency" validate:"required" example:"RUB"`
	Amount       float64 `json:"amount" validate:"required" example:"15"`
}

type ExchangeResponse struct {
	Message         string             `json:"message"`
	NewBalance      map[string]float64 `json:"new_balance" example:"USD:20.0,EUR:1.5,RUB:15.0"`
	ExchangedAmount float64            `json:"exchanged_amount" example:"15"`
}

type GetRatesResponse struct {
	Rates map[string]float64 `json:"rates" example:"USD:1.0,EUR:0.85,RUB:0.1"`
}
