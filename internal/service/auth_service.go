package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Rithvik-C18/go-flow/internal/auth"
	"github.com/Rithvik-C18/go-flow/internal/database"
	"github.com/Rithvik-C18/go-flow/internal/repository"
)

const defaultTokenTTL = 24 * time.Hour

type AuthService struct {
	repo     *repository.UserRepository
	secret   string
	tokenTTL time.Duration
}

func NewAuthService(repo *repository.UserRepository, secret string) *AuthService {
	return &AuthService{
		repo:     repo,
		secret:   secret,
		tokenTTL: defaultTokenTTL,
	}
}

func (s *AuthService) Register(username, email, password string) (*database.User, string, error) {
	passwordHash, err := auth.GeneratePasswordHash(password)
	if err != nil {
		return nil, "", err
	}

	user := &database.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := auth.GenerateToken(user, s.secret, s.tokenTTL)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(credential, password string) (*database.User, string, error) {
	user, err := s.repo.FindByUsername(credential)
	if errors.Is(err, repository.ErrNotFound) {
		user, err = s.repo.FindByEmail(credential)
	}
	if err != nil {
		return nil, "", ErrInvalidCredential
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", ErrInvalidCredential
	}

	token, err := auth.GenerateToken(user, s.secret, s.tokenTTL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) VerifyToken(tokenString string) (*auth.Claims, error) {
	return auth.ValidateToken(tokenString, s.secret)
}
