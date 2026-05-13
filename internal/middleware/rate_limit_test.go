package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewInMemoryRateLimiter(2, time.Minute) // 2 requests per minute

	r := gin.New()
	r.Use(RateLimit(limiter, func(c *gin.Context) string {
		return c.Param("source_name")
	}))
	r.POST("/webhook/:source_name", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should pass
	req1 := httptest.NewRequest("POST", "/webhook/test-source", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", w1.Code)
	}

	// Second request should pass
	req2 := httptest.NewRequest("POST", "/webhook/test-source", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("second request: expected 200, got %d", w2.Code)
	}

	// Third request should be rate limited
	req3 := httptest.NewRequest("POST", "/webhook/test-source", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("third request: expected 429, got %d", w3.Code)
	}
}

func TestRateLimitDifferentKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewInMemoryRateLimiter(1, time.Minute) // 1 request per minute per key

	r := gin.New()
	r.Use(RateLimit(limiter, func(c *gin.Context) string {
		return c.Param("source_name")
	}))
	r.POST("/webhook/:source_name", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request to source-a should pass
	req1 := httptest.NewRequest("POST", "/webhook/source-a", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("first request to source-a: expected 200, got %d", w1.Code)
	}

	// Second request to source-a should be rate limited
	req2 := httptest.NewRequest("POST", "/webhook/source-a", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request to source-a: expected 429, got %d", w2.Code)
	}

	// Request to source-b should pass (different key)
	req3 := httptest.NewRequest("POST", "/webhook/source-b", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("first request to source-b: expected 200, got %d", w3.Code)
	}
}