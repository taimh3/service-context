package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	sctx "github.com/taimaifika/service-context"
)

type customAppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *customAppError) StatusCode() int {
	return e.Code
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sv := sctx.NewServiceContext(sctx.WithName("test-service"))

	t.Run("Normal request without panic", func(t *testing.T) {
		router := gin.New()
		router.Use(Recovery(sv))
		router.GET("/ok", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Panic with CanGetStatusCode error", func(t *testing.T) {
		router := gin.New()
		router.Use(Recovery(sv))
		router.GET("/panic-app", func(c *gin.Context) {
			panic(&customAppError{Code: http.StatusForbidden, Message: "forbidden access"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic-app", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("Panic with generic error", func(t *testing.T) {
		router := gin.New()
		router.Use(Recovery(sv))
		router.GET("/panic-generic", func(c *gin.Context) {
			panic("something went completely wrong")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic-generic", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}
