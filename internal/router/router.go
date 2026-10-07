package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Rithvik-C18/go-flow/internal/handler"
	"github.com/gin-gonic/gin"
)

func NewRouter(authHandler *handler.AuthHandler, workflowHandler *handler.WorkflowHandler, authMiddleware gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	api := r.Group("/api")
	registerPublicRoutes(api, authHandler)
	registerProtectedRoutes(api, workflowHandler, authHandler, authMiddleware)
	// Keep the original API available for existing development clients.
	if os.Getenv("STATIC_DIR") == "" {
		registerPublicRoutes(r, authHandler)
		registerProtectedRoutes(r, workflowHandler, authHandler, authMiddleware)
	}
	if dir := os.Getenv("STATIC_DIR"); dir != "" {
		r.NoRoute(spaHandler(dir))
	}

	return r
}

func spaHandler(dir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || c.Request.URL.Path == "/api" || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		path := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(filepath.Clean("/"+c.Request.URL.Path), "/")))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			c.File(path)
			return
		}
		if filepath.Ext(c.Request.URL.Path) != "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.File(filepath.Join(dir, "index.html"))
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, candidate := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
			if origin != "" && origin == strings.TrimSpace(candidate) {
				allowed = true
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
