package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/config"
	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/limiter"
	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware(t *testing.T) {
	// Configura Gin para teste
	gin.SetMode(gin.TestMode)

	// Cria configuração de teste
	cfg := &config.Config{
		RateLimitIPRPS:          2, // 2 req/s para facilitar teste
		RateLimitIPBlockTime:    5 * time.Second,
		RateLimitTokenRPS:       5, // 5 req/s para token
		RateLimitTokenBlockTime: 5 * time.Second,
	}

	// Mock storage
	storage := newMockStorage()
	rl := limiter.NewRateLimiter(storage)

	// Middleware
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(rl, cfg)

	// Router de teste
	router := gin.New()
	router.Use(rateLimiterMiddleware.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("Permite requisições dentro do limite", func(t *testing.T) {
		// Primeira requisição
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "1", w.Header().Get("X-RateLimit-Remaining"))

		// Segunda requisição
		req, _ = http.NewRequest("GET", "/test", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("Bloqueia quando excede limite", func(t *testing.T) {
		// Nova instância para teste isolado
		storage := newMockStorage()
		rl := limiter.NewRateLimiter(storage)
		middleware := middleware.NewRateLimiterMiddleware(rl, cfg)

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Esgota limite (2 requisições)
		for i := 0; i < 2; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Terceira deve ser bloqueada
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "maximum number of requests")
	})

	t.Run("Token tem precedência sobre IP", func(t *testing.T) {
		storage := newMockStorage()
		rl := limiter.NewRateLimiter(storage)
		middleware := middleware.NewRateLimiterMiddleware(rl, cfg)

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Com token, deve usar limite de 5 req/s
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("API_KEY", "test-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit")) // Limite do token
		assert.Equal(t, "4", w.Header().Get("X-RateLimit-Remaining"))
	})
}
