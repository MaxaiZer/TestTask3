package services

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func Test_CreateToken_ShouldCreateValidToken(t *testing.T) {

	claims := map[string]string{"UserId": "12345", "Ip": "127.0.0.1"}
	cfg := JWTConfig{
		SecretKey: "123",
		Lifetime:  time.Minute,
		Issuer:    "MyIssuer",
		Audience:  "MyAudience",
	}

	jwtService, err := NewJwtService(cfg)
	assert.NoError(t, err)

	token, err := jwtService.CreateToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claimsFromToken, err := jwtService.ValidateToken(token)
	assert.NoError(t, err)

	assert.Equal(t, claims, claimsFromToken.ExtraClaims)
}

func Test_ValidateToken_WhenInvalidIssuer_ShouldReturnError(t *testing.T) {

	cfg := JWTConfig{
		SecretKey: "123",
		Lifetime:  time.Minute,
		Issuer:    "MyIssuer",
		Audience:  "MyAudience",
	}
	jwtService1, err := NewJwtService(cfg)
	assert.NoError(t, err)

	jwtService2 := createJwtServiceWithIssuerAndAudience("AnotherIssuer", cfg.Audience)

	token, err := jwtService2.CreateToken(map[string]string{"UserId": "12345", "Ip": "127.0.0.1"})
	assert.NoError(t, err)

	_, err = jwtService1.ValidateToken(token)
	assert.Error(t, err)
}

func Test_ValidateToken_WhenInvalidAudience_ShouldReturnError(t *testing.T) {

	cfg := JWTConfig{
		SecretKey: "123",
		Lifetime:  time.Minute,
		Issuer:    "MyIssuer",
		Audience:  "MyAudience",
	}
	jwtService1, err := NewJwtService(cfg)
	assert.NoError(t, err)

	jwtService2 := createJwtServiceWithIssuerAndAudience(cfg.Issuer, "AnotherAudience")

	token, err := jwtService2.CreateToken(map[string]string{"UserId": "12345", "Ip": "127.0.0.1"})
	assert.NoError(t, err)

	_, err = jwtService1.ValidateToken(token)
	assert.Error(t, err)
}

func Test_ValidateToken_WhenExpired_ShouldReturnError(t *testing.T) {

	cfg := JWTConfig{
		SecretKey: "123",
		Lifetime:  time.Second,
		Issuer:    "MyIssuer",
		Audience:  "MyAudience",
	}
	jwtService, err := NewJwtService(cfg)
	assert.NoError(t, err)

	token, err := jwtService.CreateToken(map[string]string{"UserId": "12345", "Ip": "127.0.0.1"})
	assert.NoError(t, err)

	time.Sleep(cfg.Lifetime + time.Second)

	_, err = jwtService.ValidateToken(token)
	assert.Error(t, err)
}

func createJwtServiceWithIssuerAndAudience(issuer string, audience string) *JwtService {

	cfg := JWTConfig{
		SecretKey: "123",
		Lifetime:  time.Minute,
		Issuer:    issuer,
		Audience:  audience,
	}
	jwt, err := NewJwtService(cfg)

	if err != nil {
		log.Fatalf("failed to create jwt service: %v", err)
	}

	return jwt
}
