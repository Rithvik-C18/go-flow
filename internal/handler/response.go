package handler

import (
	"errors"
	"net/http"

	"github.com/Rithvik-C18/go-flow/internal/repository"
	"github.com/Rithvik-C18/go-flow/internal/service"
	"github.com/gin-gonic/gin"
)

func respondJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, repository.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, repository.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, service.ErrValidation):
		status = http.StatusBadRequest
	}

	respondJSON(c, status, gin.H{"error": err.Error()})
}

func respondValidationError(c *gin.Context, message string) {
	respondJSON(c, http.StatusBadRequest, gin.H{"error": message})
}
