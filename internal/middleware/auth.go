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
			refreshToken, err := c.Cookie("refresh_token")
			if err != nil || refreshToken == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
				return
			}

			newAccessToken, newRefreshToken, err := authService.RefreshAccessToken(refreshToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
				return
			}

			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie(
				"refresh_token",
				newRefreshToken,
				int(7*24*60*60),
				"/",
				"",
				true,
				true,
			)

			c.Header("New-Access-Token", newAccessToken)

			claims, err = authService.VerifyToken(newAccessToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "failed to verify new token"})
				return
			}
		}

		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
