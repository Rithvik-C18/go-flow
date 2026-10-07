package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProductionRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("app shell"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STATIC_DIR", dir)
	r := NewRouter(nil, nil, func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) })
	for _, tc := range []struct {
		path   string
		status int
		body   string
	}{
		{"/", 200, "app shell"},
		{"/auth", 200, "app shell"},
		{"/workflows/example", 200, "app shell"},
		{"/api/workflows", 401, ""},
		{"/api/missing", 404, ""},
		{"/missing.js", 404, ""},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d", w.Code, tc.status)
			}
			if tc.body != "" && w.Body.String() != tc.body {
				t.Fatalf("unexpected body %q", w.Body.String())
			}
		})
	}
}

func TestCORSRejectsUnlistedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ALLOWED_ORIGINS", "https://app.example")
	r := gin.New()
	r.Use(corsMiddleware())
	for _, origin := range []string{"https://app.example", "https://attacker.example"} {
		req := httptest.NewRequest("OPTIONS", "/api/workflows", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		expected := ""
		if origin == "https://app.example" {
			expected = origin
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != expected {
			t.Fatalf("origin %q allowed as %q", origin, got)
		}
	}
}
