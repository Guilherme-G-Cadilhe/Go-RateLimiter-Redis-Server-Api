package limiter

import (
	"context"
	"fmt"
	"time"
)

type RateLimiter struct {
	storage StorageStrategy
}

type LimitConfig struct {
	RPS       int           // Requests per second
	BlockTime time.Duration // Tempo de bloqueio quando excedido
}

type CheckResult struct {
	Allowed   bool
	Remaining int
	ResetTime time.Time
	Blocked   bool
}

func NewRateLimiter(storage StorageStrategy) *RateLimiter {
	return &RateLimiter{
		storage: storage,
	}
}

func (rl *RateLimiter) Check(ctx context.Context, key string, config LimitConfig) (*CheckResult, error) {
	// Verifica se está bloqueado
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar bloqueio: %w", err)
	}

	if blocked {
		return &CheckResult{
			Allowed: false,
			Blocked: true,
		}, nil
	}

	// Chave para contagem de requisições
	countKey := fmt.Sprintf("rate:%s", key)

	// Incrementa contador com TTL de 1 segundo
	count, existed, err := rl.storage.Increment(ctx, countKey, time.Second)
	if err != nil {
		return nil, fmt.Errorf("erro ao incrementar contador: %w", err)
	}

	// Se não existia antes, é a primeira requisição neste segundo
	if !existed {
		count = 1
	}

	// Calcula informações de reset
	resetTime := time.Now().Add(time.Second)
	remaining := config.RPS - count
	if remaining < 0 {
		remaining = 0
	}

	// Verifica se excedeu o limite
	if count > config.RPS {
		// Bloqueia por BlockTime
		if err := rl.storage.Block(ctx, key, config.BlockTime); err != nil {
			return nil, fmt.Errorf("erro ao bloquear chave: %w", err)
		}

		return &CheckResult{
			Allowed:   false,
			Remaining: 0,
			ResetTime: resetTime,
			Blocked:   false, // Acabou de ser bloqueado
		}, nil
	}

	return &CheckResult{
		Allowed:   true,
		Remaining: remaining,
		ResetTime: resetTime,
		Blocked:   false,
	}, nil
}
