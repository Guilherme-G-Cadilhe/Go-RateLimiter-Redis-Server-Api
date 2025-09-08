package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/config"
	"github.com/gin-gonic/gin"

	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/limiter"
)

type RateLimiterMiddleware struct {
	limiter *limiter.RateLimiter
	config  *config.Config
}

func NewRateLimiterMiddleware(rateLimiter *limiter.RateLimiter, cfg *config.Config) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: rateLimiter,
		config:  cfg,
	}
}

// Middleware retorna a função middleware do Gin
func (rlm *RateLimiterMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1. Extrair IP do cliente
		clientIP := getClientIP(c)

		// 2. Verificar se existe token de API
		apiToken := c.GetHeader("API_KEY")

		var key string
		var limitConfig limiter.LimitConfig

		// 3. Determinar qual limite usar (Token sobrepõe IP)
		if apiToken != "" {
			// Usa configuração do token (mais permissiva)
			key = fmt.Sprintf("token:%s", apiToken)
			limitConfig = limiter.LimitConfig{
				RPS:       rlm.config.RateLimitTokenRPS,
				BlockTime: rlm.config.RateLimitTokenBlockTime,
			}
		} else {
			// Usa configuração do IP
			key = fmt.Sprintf("ip:%s", clientIP)
			limitConfig = limiter.LimitConfig{
				RPS:       rlm.config.RateLimitIPRPS,
				BlockTime: rlm.config.RateLimitIPBlockTime,
			}
		}

		// 4. Verificar rate limit
		result, err := rlm.limiter.Check(ctx, key, limitConfig)
		if err != nil {
			// Em caso de erro no Redis/storage, logamos mas não bloqueamos
			// Isso evita que problemas no Redis derrubem a aplicação
			fmt.Printf("Erro no rate limiter: %v\n", err)
			c.Next() // Continua sem limitação
			return
		}

		// 5. Adicionar headers informativos (mesmo quando permitido)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limitConfig.RPS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))

		// 6. Verificar se deve bloquear
		if !result.Allowed {
			// Headers adicionais para requisições bloqueadas
			c.Header("Retry-After", fmt.Sprintf("%d", int(limitConfig.BlockTime.Seconds())))

			// Resposta HTTP 429 - Too Many Requests
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "you have reached the maximum number of requests or actions allowed within a certain time frame",
				"retry_after_seconds": int(limitConfig.BlockTime.Seconds()),
			})

			// Aborta a execução - não chama os próximos handlers
			c.Abort()
			return
		}

		// 7. Se chegou aqui, está dentro do limite - continua
		c.Next()
	}
}

// getClientIP extrai o IP real do cliente considerando proxies/load balancers
func getClientIP(c *gin.Context) string {
	// 1. Verifica header X-Forwarded-For (comum em load balancers)
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// Pode ter múltiplos IPs separados por vírgula
		// O primeiro é geralmente o IP original do cliente
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 2. Verifica header X-Real-IP (comum em nginx)
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		if net.ParseIP(xRealIP) != nil {
			return xRealIP
		}
	}

	// 3. Fallback para RemoteAddr (IP direto)
	// Remove a porta se existir (formato IP:porta)
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		// Se não conseguir fazer split, provavelmente é só o IP
		return c.Request.RemoteAddr
	}

	return ip
}
