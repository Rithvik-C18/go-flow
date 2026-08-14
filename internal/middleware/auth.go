package middleware

import (
	"net/http"
	"strings"

	"github.com/Rithvik-C18/go-flow/internal/service"
	"github.com/gin-gonic/gin"
)

func Auth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := authService.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
