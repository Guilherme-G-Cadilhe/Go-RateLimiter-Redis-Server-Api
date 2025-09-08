#!/bin/bash
# Script para ambiente de produção

echo "🚀 Iniciando ambiente de produção..."

# Para containers se estiverem rodando
docker-compose down

# Rebuild e inicia todos os serviços
docker-compose up --build -d

echo "⏳ Aguardando serviços inicializarem..."
sleep 10

# Mostra status dos containers
docker-compose ps

echo "✅ Aplicação disponível em: http://localhost:8080"
echo "🔍 Para ver logs: docker-compose logs -f"
