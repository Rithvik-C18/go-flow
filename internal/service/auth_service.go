package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Rithvik-C18/go-flow/internal/auth"
	"github.com/Rithvik-C18/go-flow/internal/config"
	"github.com/Rithvik-C18/go-flow/internal/database"
	"github.com/Rithvik-C18/go-flow/internal/repository"
)

type AuthService struct {
	repo              *repository.UserRepository
	cfg               config.Config
}

func NewAuthService(repo *repository.UserRepository, cfg config.Config) *AuthService {
	return &AuthService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *AuthService) Register(username, email, password string) (*database.User, string, string, error) {
	passwordHash, err := auth.GeneratePasswordHash(password)
	if err != nil {
		return nil, "", "", err
	}

	user := &database.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, "", "", err
	}

	accessToken, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(credential, password string) (*database.User, string, string, error) {
	user, err := s.repo.FindByUsername(credential)
	if errors.Is(err, repository.ErrNotFound) {
		user, err = s.repo.FindByEmail(credential)
	}
	if err != nil {
		return nil, "", "", ErrInvalidCredential
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", "", ErrInvalidCredential
	}

	accessToken, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) VerifyToken(tokenString string) (*auth.Claims, error) {
	return auth.ValidateToken(tokenString, s.cfg.JWT.Secret)
}

func (s *AuthService) generateTokenPair(user *database.User) (string, string, error) {
	accessToken, err := auth.GenerateAccessToken(user, s.cfg.JWT.Secret, s.cfg.JWT.AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user, s.cfg.JWT.Secret, s.cfg.JWT.RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshTokenHash := auth.HashRefreshToken(refreshToken)
	expiresAt := time.Now().Add(s.cfg.JWT.RefreshTokenTTL)
	user.RefreshTokenHash = refreshTokenHash
	user.RefreshTokenExpiresAt = &expiresAt

	if err := s.repo.Update(user); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshAccessToken(refreshToken string) (string, string, error) {
	claims, err := auth.ValidateRefreshToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	refreshTokenHash := auth.HashRefreshToken(refreshToken)
	if user.RefreshTokenHash != refreshTokenHash {
		return "", "", errors.New("refresh token does not match")
	}

	if user.RefreshTokenExpiresAt != nil && time.Now().After(*user.RefreshTokenExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	newAccessToken, newRefreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new token pair: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(userID uint) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.RefreshTokenHash = ""
	user.RefreshTokenExpiresAt = nil

	return s.repo.Update(user)
}
