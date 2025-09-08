#!/bin/bash
# Script para desenvolvimento local

echo "🐳 Iniciando ambiente de desenvolvimento..."

# Para containers se estiverem rodando
docker-compose down

# Sobe apenas o Redis para desenvolvimento local
docker-compose up -d redis

echo "⏳ Aguardando Redis inicializar..."
sleep 5

echo "✅ Redis pronto! Agora execute: go run cmd/server/main.go"
echo "🔍 Redis disponível em: localhost:6379"
echo "📊 Para monitorar Redis: docker exec -it rate_limiter_redis redis-cli monitor"
