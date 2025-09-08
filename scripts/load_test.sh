#!/bin/bash

echo "🔥 Iniciando testes de carga do Rate Limiter"

# Verifica se aplicação está rodando
if ! curl -s http://localhost:8080/ > /dev/null; then
    echo "❌ Aplicação não está rodando na porta 8080"
    echo "Execute: make run ou make prod"
    exit 1
fi

echo "📊 Teste 1: Rate limit por IP (5 req/s)"
echo "Enviando 20 requisições rápidas..."

# Teste básico com curl
for i in {1..10}; do
    echo -n "Req $i: "
    curl -s -w "%{http_code}\n" -o /dev/null http://localhost:8080/test
    sleep 0.1
done

echo -e "\n🔑 Teste 2: Rate limit com Token (100 req/s)"
echo "Enviando 10 requisições com token..."

for i in {1..10}; do
    echo -n "Req $i: "
    curl -s -w "%{http_code}\n" -o /dev/null -H "API_KEY: test-token-123" http://localhost:8080/test
    sleep 0.05
done

echo -e "\n🚀 Teste 3: Simulando múltiplos IPs"
echo "Usando diferentes X-Forwarded-For..."

IPS=("192.168.1.1" "192.168.1.2" "192.168.1.3" "10.0.0.1" "10.0.0.2")

for ip in "${IPS[@]}"; do
    echo -n "IP $ip: "
    for i in {1..3}; do
        curl -s -w "%{http_code} " -o /dev/null -H "X-Forwarded-For: $ip" http://localhost:8080/test
    done
    echo
done

echo -e "\n✅ Testes concluídos!"
echo "💡 Para testes mais avançados, instale 'hey': go install github.com/rakyll/hey@latest"
echo "   Depois execute: make load-test"
