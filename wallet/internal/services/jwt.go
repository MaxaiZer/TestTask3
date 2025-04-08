package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SecretKey string
	Lifetime  time.Duration
	Issuer    string
	Audience  string
}

type CustomClaims struct {
	jwt.RegisteredClaims
	ExtraClaims map[string]string `json:"extra,omitempty"`
}

type JwtService struct {
	secretKey          string
	issuer             string
	audience           string
	lifetime           time.Duration
	signingMethod      jwt.SigningMethod
	refreshTokenLength int
}

func NewJwtService(cfg JWTConfig) (*JwtService, error) {

	return &JwtService{
		secretKey:          cfg.SecretKey,
		signingMethod:      jwt.SigningMethodHS512,
		lifetime:           cfg.Lifetime,
		issuer:             cfg.Issuer,
		audience:           cfg.Audience,
		refreshTokenLength: 32,
	}, nil
}

func (s *JwtService) CreateToken(extraClaims map[string]string) (string, error) {
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.lifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		ExtraClaims: extraClaims,
	}

	token := jwt.NewWithClaims(s.signingMethod, claims)

	signedToken, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *JwtService) ValidateToken(token string) (*CustomClaims, error) {

	parsedToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (any, error) {

		if token.Method != s.signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*CustomClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if s.issuer != claims.Issuer || len(claims.Audience) == 0 || s.audience != claims.Audience[0] {
		return nil, fmt.Errorf("invalid audience or issuer")
	}

	return claims, nil
}
