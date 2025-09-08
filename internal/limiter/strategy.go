package limiter

import (
	"context"
	"time"
)

// StorageStrategy define a interface para diferentes tipos de storage
type StorageStrategy interface {
	// Get retorna o número atual de tokens para a chave
	Get(ctx context.Context, key string) (int, error)

	// Set define o número de tokens com TTL
	Set(ctx context.Context, key string, tokens int, ttl time.Duration) error

	// Increment incrementa o contador de tokens atomicamente
	// Retorna o novo valor e se a chave já existia
	Increment(ctx context.Context, key string, ttl time.Duration) (int, bool, error)

	// IsBlocked verifica se a chave está bloqueada
	IsBlocked(ctx context.Context, key string) (bool, error)

	// Block bloqueia uma chave por um período
	Block(ctx context.Context, key string, blockTime time.Duration) error
}
