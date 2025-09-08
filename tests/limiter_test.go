package tests

import (
	"context"
	"testing"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-RateLimiter-Redis-Server-Api/internal/limiter"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock strategy para testes unitários (sem Redis)
type mockStorage struct {
	data    map[string]int
	blocked map[string]bool
	ttls    map[string]time.Time
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		data:    make(map[string]int),
		blocked: make(map[string]bool),
		ttls:    make(map[string]time.Time),
	}
}

func (m *mockStorage) Get(ctx context.Context, key string) (int, error) {
	// Verifica se TTL expirou
	if ttl, exists := m.ttls[key]; exists && time.Now().After(ttl) {
		delete(m.data, key)
		delete(m.ttls, key)
		return 0, nil
	}

	if val, exists := m.data[key]; exists {
		return val, nil
	}
	return 0, nil
}

func (m *mockStorage) Set(ctx context.Context, key string, tokens int, ttl time.Duration) error {
	m.data[key] = tokens
	m.ttls[key] = time.Now().Add(ttl)
	return nil
}

func (m *mockStorage) Increment(ctx context.Context, key string, ttl time.Duration) (int, bool, error) {
	existed := false
	if _, exists := m.data[key]; exists {
		existed = true
	}

	m.data[key]++
	m.ttls[key] = time.Now().Add(ttl)

	return m.data[key], existed, nil
}

func (m *mockStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	return m.blocked[key], nil
}

func (m *mockStorage) Block(ctx context.Context, key string, blockTime time.Duration) error {
	m.blocked[key] = true

	// Remove bloqueio após blockTime (simulação)
	time.AfterFunc(blockTime, func() {
		delete(m.blocked, key)
	})

	return nil
}

// Teste unitário básico do rate limiter
func TestRateLimiter_Basic(t *testing.T) {
	storage := newMockStorage()
	rl := limiter.NewRateLimiter(storage)

	config := limiter.LimitConfig{
		RPS:       5, // 5 requisições por segundo
		BlockTime: 10 * time.Second,
	}

	ctx := context.Background()
	key := "test-ip"

	// Primeiras 5 requisições devem passar
	for i := 1; i <= 5; i++ {
		result, err := rl.Check(ctx, key, config)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Requisição %d deveria ser permitida", i)
		assert.Equal(t, 5-i, result.Remaining, "Remaining incorreto na requisição %d", i)
		assert.False(t, result.Blocked)
	}

	// 6ª requisição deve ser bloqueada
	result, err := rl.Check(ctx, key, config)
	require.NoError(t, err)
	assert.False(t, result.Allowed, "6ª requisição deveria ser bloqueada")
	assert.Equal(t, 0, result.Remaining)

	// Próxima verificação deve mostrar como bloqueado
	result, err = rl.Check(ctx, key, config)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.True(t, result.Blocked, "Deveria estar bloqueado")
}

// Teste com Redis real (integração)
func TestRateLimiter_Redis(t *testing.T) {
	// Conecta ao Redis de teste
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Database de teste
	})

	// Testa conexão
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis não disponível para teste de integração")
	}

	defer rdb.Close()

	// Limpa database de teste
	rdb.FlushDB(ctx)

	// Cria strategy e limiter
	strategy := limiter.NewRedisStrategy(rdb)
	rl := limiter.NewRateLimiter(strategy)

	config := limiter.LimitConfig{
		RPS:       3,
		BlockTime: 5 * time.Second,
	}

	key := "test-redis-key"

	// Testa sequência de requisições
	for i := 1; i <= 3; i++ {
		result, err := rl.Check(ctx, key, config)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Requisição %d deveria passar", i)
	}

	// 4ª deve bloquear
	result, err := rl.Check(ctx, key, config)
	require.NoError(t, err)
	assert.False(t, result.Allowed, "4ª requisição deveria bloquear")

	// Verifica se está bloqueado
	blocked, err := strategy.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.True(t, blocked, "Deveria estar bloqueado no Redis")

	// Limpa para próximos testes
	rdb.FlushDB(ctx)
}

// Teste de concorrência (race conditions)
func TestRateLimiter_Concurrency(t *testing.T) {
	storage := newMockStorage()
	rl := limiter.NewRateLimiter(storage)

	config := limiter.LimitConfig{
		RPS:       10,
		BlockTime: 5 * time.Second,
	}

	ctx := context.Background()
	key := "concurrent-test"

	// Canal para coletar resultados
	results := make(chan bool, 20)

	// Dispara 20 goroutines simultâneas
	for i := 0; i < 20; i++ {
		go func() {
			result, err := rl.Check(ctx, key, config)
			if err != nil {
				results <- false
				return
			}
			results <- result.Allowed
		}()
	}

	// Coleta resultados
	allowed := 0
	blocked := 0

	for i := 0; i < 20; i++ {
		if <-results {
			allowed++
		} else {
			blocked++
		}
	}

	// Deve permitir exatamente 10 (RPS) e bloquear 10
	assert.Equal(t, 10, allowed, "Deveria permitir exatamente 10 requisições")
	assert.Equal(t, 10, blocked, "Deveria bloquear exatamente 10 requisições")
}
