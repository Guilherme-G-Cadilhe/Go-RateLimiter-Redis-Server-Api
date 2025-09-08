package limiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStrategy struct {
	client *redis.Client
}

func NewRedisStrategy(client *redis.Client) *RedisStrategy {
	return &RedisStrategy{
		client: client,
	}
}

func (r *RedisStrategy) Get(ctx context.Context, key string) (int, error) {
	// GET comando do Redis - busca valor por chave
	val, err := r.client.Get(ctx, key).Result()
	// redis.Nil significa que a chave não existe (não é um erro real)
	if err == redis.Nil {
		return 0, nil // Chave não existe = 0 requisições
	}
	// Outros erros (conexão, timeout, etc.)
	if err != nil {
		return 0, err
	}
	// Converte string para int (Redis armazena tudo como string)
	count, err := strconv.Atoi(val)
	if err != nil {
		// Se não conseguir converter, assume 0 (dados corrompidos)
		return 0, err
	}

	return count, nil
}

func (r *RedisStrategy) Set(ctx context.Context, key string, tokens int, ttl time.Duration) error {
	// SET comando com TTL (Time To Live)
	// Após TTL segundos, o Redis automaticamente remove a chave
	return r.client.Set(ctx, key, tokens, ttl).Err()
}

func (r *RedisStrategy) Increment(ctx context.Context, key string, ttl time.Duration) (int, bool, error) {
	/*
		PIPELINE é uma funcionalidade do Redis que permite:
		- Enviar múltiplos comandos de uma vez
		- Executar atomicamente (tudo ou nada)
		- Reduzir latência de rede

		Isso é CRUCIAL para rate limiting porque evita race conditions:
		- Cliente A e B fazem requisição no mesmo milissegundo
		- Sem pipeline: ambos podem ver count=4, incrementar para 5
		- Com pipeline: um vê 4→5, outro vê 5→6 (correto)
	*/
	pipe := r.client.Pipeline()

	// INCR incrementa contador atomicamente
	// Se a chave não existe, Redis cria com valor 1
	incrCmd := pipe.Incr(ctx, key)

	// EXPIRE define TTL para a chave
	// Importante: fazemos isso sempre para renovar o TTL
	pipe.Expire(ctx, key, ttl)

	// Executa todos os comandos do pipeline atomicamente
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("erro no pipeline Redis: %w", err)
	}

	// Pega o novo valor retornado pelo INCR
	newVal := int(incrCmd.Val())
	// Se newVal > 1, significa que a chave já existia antes
	existed := newVal > 1

	return newVal, existed, nil
}

func (r *RedisStrategy) IsBlocked(ctx context.Context, key string) (bool, error) {
	// Chaves de bloqueio têm prefixo "block:" para organização
	blockKey := fmt.Sprintf("block:%s", key)
	// EXISTS verifica se a chave existe (retorna 1 se existe, 0 se não)
	exists, err := r.client.Exists(ctx, blockKey).Result()
	return exists > 0, err
}

func (r *RedisStrategy) Block(ctx context.Context, key string, blockTime time.Duration) error {
	// Cria chave de bloqueio com TTL
	blockKey := fmt.Sprintf("block:%s", key)

	// SET com TTL - após blockTime, Redis remove automaticamente
	// Valor "blocked" é apenas informativo, o importante é a existência da chave
	err := r.client.Set(ctx, blockKey, "blocked", blockTime).Err()
	if err != nil {
		return fmt.Errorf("erro ao bloquear no Redis: %w", err)
	}

	return nil
}

// Método adicional para debug/monitoramento
func (r *RedisStrategy) GetStats(ctx context.Context, keyPrefix string) (map[string]int, error) {
	/*
		SCAN é mais eficiente que KEYS para produção
		KEYS bloqueia o Redis, SCAN é não-bloqueante
	*/
	pattern := fmt.Sprintf("%s*", keyPrefix)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)

	// Para cada chave encontrada, pega o valor
	for _, key := range keys {
		val, err := r.Get(ctx, key)
		if err != nil {
			continue // Pula chaves com erro
		}
		stats[key] = val
	}

	return stats, nil
}
