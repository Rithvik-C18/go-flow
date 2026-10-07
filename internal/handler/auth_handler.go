package handler

import (
	"errors"
	"net/http"

	"github.com/Rithvik-C18/go-flow/internal/database"
	"github.com/Rithvik-C18/go-flow/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "username, email and password are required")
		return
	}

	user, accessToken, refreshToken, err := h.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, refreshToken)

	respondJSON(c, http.StatusCreated, gin.H{
		"access_token": accessToken,
		"user":         userResponse(user),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "username and password are required")
		return
	}

	user, accessToken, refreshToken, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			respondJSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		respondError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, refreshToken)

	respondJSON(c, http.StatusOK, gin.H{
		"access_token": accessToken,
		"user":         userResponse(user),
	})
}

func userResponse(user *database.User) gin.H {
	return gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	}
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		respondJSON(c, http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	accessToken, newRefreshToken, err := h.service.RefreshAccessToken(refreshToken)
	if err != nil {
		respondJSON(c, http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	h.setRefreshTokenCookie(c, newRefreshToken)

	respondJSON(c, http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		respondJSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		respondJSON(c, http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.service.Logout(userIDUint); err != nil {
		respondError(c, err)
		return
	}

	h.clearRefreshTokenCookie(c)

	respondJSON(c, http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		token,
		int(7*24*60*60), // 7 days
		"/",
		"",
		true,  // secure
		true,  // httpOnly
	)
}

func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)
}
