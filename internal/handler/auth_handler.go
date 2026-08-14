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

	user, token, err := h.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	respondJSON(c, http.StatusCreated, gin.H{
		"token": token,
		"user":  userResponse(user),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "username and password are required")
		return
	}

	user, token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			respondJSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		respondError(c, err)
		return
	}

	respondJSON(c, http.StatusOK, gin.H{
		"token": token,
		"user":  userResponse(user),
	})
}

func userResponse(user *database.User) gin.H {
	return gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	}
}
