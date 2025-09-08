package main

import (
	"fmt"
	"log"

	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/config"
	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/limiter"
	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/middleware"
	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Carrega configurações do .env
	cfg := config.LoadConfig()

	// 2. Conecta ao Redis
	redisClient, err := storage.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar Redis: %v", err)
	}
	defer redisClient.Close() // Fecha conexão ao terminar

	// 3. Cria strategy e rate limiter
	redisStrategy := limiter.NewRedisStrategy(redisClient)
	rateLimiter := limiter.NewRateLimiter(redisStrategy)

	// 4. Cria middleware
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(rateLimiter, cfg)

	// 5. Configura Gin router
	router := gin.Default()

	// Aplica middleware de rate limiting globalmente
	router.Use(rateLimiterMiddleware.Middleware())

	// 6. Define rotas de exemplo
	setupRoutes(router)

	// 7. Inicia servidor
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("🚀 Servidor iniciando na porta %s\n", cfg.ServerPort)
	fmt.Printf("📊 Rate Limit IP: %d req/s\n", cfg.RateLimitIPRPS)
	fmt.Printf("🔑 Rate Limit Token: %d req/s\n", cfg.RateLimitTokenRPS)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func setupRoutes(router *gin.Engine) {
	// Rota simples para teste
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Rate Limiter funcionando!",
			"ip":      c.ClientIP(),
			"token":   c.GetHeader("API_KEY"),
		})
	})

	// Rota para simular carga
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Teste de carga",
			"headers": c.Request.Header,
		})
	})

	// Rota protegida por token
	router.GET("/protected", func(c *gin.Context) {
		token := c.GetHeader("API_KEY")
		if token == "" {
			c.JSON(401, gin.H{"error": "Token obrigatório"})
			return
		}

		c.JSON(200, gin.H{
			"message": "Acesso autorizado",
			"token":   token,
		})
	})

	// Rota para estatísticas (debug)
	router.GET("/stats", func(c *gin.Context) {
		// Aqui você poderia implementar endpoint para ver estatísticas
		c.JSON(200, gin.H{
			"message": "Endpoint de estatísticas - implementar se necessário",
		})
	})
}
