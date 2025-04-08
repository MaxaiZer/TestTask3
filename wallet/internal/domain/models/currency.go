package models

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	RUB Currency = "RUB"
)

func (c Currency) IsValid() bool {
	switch c {
	case USD, EUR, RUB:
		return true
	default:
		return false
	}
}
