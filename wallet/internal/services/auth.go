package services

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	errs "test-task/wallet/internal/domain/errors"
	"test-task/wallet/internal/domain/models"
)

type AuthRepository interface {
	AddUser(ctx context.Context, name string, password []byte, email string) error
	GetUserByName(ctx context.Context, name string) (*models.User, error)
}

type AuthService struct {
	jwt  *JwtService
	repo AuthRepository
}

func NewAuthService(jwt *JwtService, repo AuthRepository) *AuthService {
	return &AuthService{jwt: jwt, repo: repo}
}

func (a *AuthService) Register(ctx context.Context, name, password, email string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate hash from password: %w", err)
	}
	return a.repo.AddUser(ctx, name, hashedPassword, email)
}

func (a *AuthService) Login(ctx context.Context, name, password string) (string, error) {

	user, err := a.repo.GetUserByName(ctx, name)
	if err != nil {
		if errors.Is(err, errs.UserNotExists) {
			return "", err
		}
		return "", fmt.Errorf("failed to get user by name: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return "", errs.WrongPassword
	}

	token, err := a.jwt.CreateToken(map[string]string{"id": strconv.FormatInt(user.ID, 10)})
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}
	return token, nil
}
