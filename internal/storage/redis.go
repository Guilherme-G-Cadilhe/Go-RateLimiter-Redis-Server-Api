package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient cria uma nova conexão com Redis
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	// Configuração do cliente Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort), // localhost:6379
		Password: cfg.RedisPassword,                                  // sem senha por padrão
		DB:       cfg.RedisDB,                                        // database 0 por padrão

		// Configurações de pool de conexões
		PoolSize:     10,              // 10 conexões simultâneas máximo
		MinIdleConns: 2,               // mínimo 2 conexões idle
		MaxRetries:   3,               // 3 tentativas em caso de erro
		DialTimeout:  5 * time.Second, // timeout para conectar
		ReadTimeout:  3 * time.Second, // timeout para ler
		WriteTimeout: 3 * time.Second, // timeout para escrever
	})

	// Testa a conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PING é o comando mais simples para testar conectividade
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar com Redis: %w", err)
	}

	fmt.Println("✅ Conectado ao Redis com sucesso!")
	return rdb, nil
}
