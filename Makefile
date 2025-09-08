# Makefile para automatizar tarefas comuns

.PHONY: dev prod test clean logs

# Desenvolvimento local (apenas Redis no Docker)
dev:
	@chmod +x scripts/dev.sh
	@./scripts/dev.sh

# Produção (tudo no Docker)
prod:
	@chmod +x scripts/prod.sh
	@./scripts/prod.sh

# Executa aplicação local
run:
	go run cmd/server/main.go

# Executa testes
test:
	go test ./... -v -cover

# Testes de carga (precisa instalar hey: go install github.com/rakyll/hey@latest)
load-test:
	@echo "🔥 Teste de carga - IP (5 req/s limite)"
	hey -n 20 -c 5 http://localhost:8080/test
	@echo "\n🔑 Teste de carga - Token (100 req/s limite)"  
	hey -n 200 -c 10 -H "API_KEY: test-token-123" http://localhost:8080/test

# Para todos os containers
stop:
	docker-compose down

# Remove containers e volumes (reset completo)
clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

# Monitora logs em tempo real
logs:
	docker-compose logs -f

# Acessa Redis CLI
redis-cli:
	docker exec -it rate_limiter_redis redis-cli

# Estatísticas do Redis
redis-stats:
	docker exec -it rate_limiter_redis redis-cli info stats
