# 🚀 Go Rate Limiter - Redis & Gin Framework

Sistema de rate limiting distribuído em Go, implementando Strategy Pattern e operações atômicas com Redis. Controla tráfego de requisições por IP ou token com middleware reutilizável para alta performance e escalabilidade.

## 🏗️ Arquitetura

```
Cliente → Middleware → Rate Limiter → Redis Strategy → Redis
       ↓
Token > IP Priority
       ↓
Atomic Operations (Pipeline)
```

**Componentes principais**:

- **Middleware**: Integração transparente com Gin framework
- **Rate Limiter:** Lógica core com algoritmo Token Bucket
- **Strategy Pattern**: Abstração para diferentes storages (Redis/Memória/DB)
- **Redis Storage:** Operações atômicas com pipeline e TTL automático
- **Config Manager:** Configuração via Viper (.env + variáveis ambiente)

**Sistema de Concorrência:** -**Pipeline Redis:** Operações atômicas (INCR + EXPIRE) para evitar race conditions -**Connection Pool:** Pool otimizado de conexões Redis -**Graceful Degradation:** Sistema continua funcionando mesmo com falha no Redis -**TTL Automático:** Redis gerencia expiração de chaves automaticamente

## 🚀 Como Executar

### Pré-requisitos

- Docker e Docker Compose instalados
- Opcionalmente: Go 1.23+ para desenvolvimento local

### Executar o sistema completo

```bash
# Clone o repositório
git clone <repo>
cd rate-limiter
# Inicia todos os serviços (Redis + Aplicação)
make prod
# Para parar
make stop
```

### Executar localmente (desenvolvimento)

```bash
# Instalar dependências
go mod tidy
# Subir apenas Redis via Docker
make dev
# Executar aplicação
make run
```

### ⚡ Sistema de Rate Limiting

**Algoritmo Token Bucket**
O sistema implementa Token Bucket com Redis para máxima eficiência:

**Funcionamento:**

- Cada IP/token possui um "bucket" de tokens no Redis
- Cada requisição consome 1 token
- Tokens são repostos automaticamente via TTL (1 segundo)
- Sem tokens = bloqueio temporal configurável

**Precedência Token > IP**

```bash
# Limite por IP: 10 req/s
curl http://localhost:8080/test

# Limite por Token: 100 req/s (sobrepõe o IP)
curl -H "API_KEY: token123" http://localhost:8080/test
```

**Operações Atômicas Redis**

- **Pipeline:** INCR + EXPIRE em operação única
- **Race Condition Safe:** Múltiplas instâncias podem usar mesmo Redis
- **TTL Automático:** Cleanup automático de chaves expiradas
- **Bloqueio Temporal:** Chaves block:\* com TTL configurável

## 📁 Estrutura do Projeto

```
Go-RateLimiter/
├── cmd/server/ # Aplicação principal
│ └── main.go
├── internal/
│ ├── config/ # Configuração via Viper
│ │ └── config.go
│ ├── limiter/ # Core rate limiting
│ │ ├── limiter.go # ← Lógica principal
│ │ ├── strategy.go # ← Interface Strategy
│ │ └── redis_strategy.go # ← Implementação Redis
│ ├── middleware/ # Integração Gin
│ │ └── rate_limiter.go # ← Middleware + IP extraction
│ └── storage/ # Storage clients
│ └── redis.go # ← Cliente Redis otimizado
├── tests/ # Testes automatizados
│ ├── limiter_test.go # ← Testes unitários + integração
│ └── middleware_test.go # ← Testes middleware Gin
├── scripts/ # Scripts utilitários
│ ├── dev.sh
│ ├── prod.sh
│ └── load_test.sh
├── docker-compose.yml # Orquestração containers
├── Dockerfile # Multi-stage build
├── Makefile # Automação comandos
├── .env # Configurações
└── go.mod
```

## 🧪 Testes e Validação

**Testes Automatizados**

```bash
# Executar todos os testes
make test
# Teste de carga manual
./scripts/load_test.sh
# Teste com hey (se instalado)
make load-test
```

**Cenários Testados**

- **Rate Limiting Básico:** Permite até limite, bloqueia excesso
- **Precedência Token:** Token sobrepõe configuração de IP
- **Concorrência:** 20 goroutines simultâneas sem race conditions
- **Integração Redis:** Testes com Redis real funcionando
- **Middleware Gin:** Headers corretos e integração transparente

## 🔧 Configuração

**Variáveis de Ambiente (.env)**

```
# Rate Limits
RATE_LIMIT_IP_RPS=10
RATE_LIMIT_IP_BLOCK_TIME=300s
RATE_LIMIT_TOKEN_RPS=100
RATE_LIMIT_TOKEN_BLOCK_TIME=600s

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Server
SERVER_PORT=8080
```

### Headers e Respostas

**Headers enviados pelo sistema:**

- X-RateLimit-Limit: Limite por segundo
- X-RateLimit-Remaining: Requisições restantes
- X-RateLimit-Reset: Timestamp do reset
- Retry-After: Segundos para tentar novamente (quando bloqueado)

**Resposta HTTP 429:**

```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
  "retry_after_seconds": 300
}
```

## 🧩 Conceitos Implementados

**Strategy Pattern**

- **Interface comum** para diferentes storages
- **Redis Strategy** implementada com operações atômicas
- **Mock Strategy** para testes unitários
- **Facilita extensão** para PostgreSQL, MongoDB, etc.

**Middleware Pattern**

- **Integração transparente** com qualquer aplicação Gin
- **Extração inteligente de IP** (considera proxies/load balancers)
- **Headers informativos** para debugging
- **Graceful degradation** em caso de falhas

**Patterns de Design**

- **Strategy Pattern** para abstração de storage
- **Pipeline Pattern** para operações Redis atômicas
- **Middleware Pattern** para integração com frameworks
- **Dependency Injection** para baixo acoplamento

**Performance e Escalabilidade**

- **Connection Pool** Redis otimizado (10 conexões)
- **Operações atômicas** previnem race conditions
- **TTL automático** evita memory leaks
- **Horizontal scaling** com Redis compartilhado

### 🐳 Docker e Produção

**Multi-stage Build**

- **Stage 1:** Compilação com Go completo
- **Stage 2:** Runtime com Alpine Linux (~20MB final)
- **Security:** Executa como usuário não-root
- **Health Checks:** Verificação automática de saúde dos serviços

Comandos Úteis

```bash
make prod # Deploy produção completo
make dev # Desenvolvimento (só Redis)
make logs # Ver logs em tempo real
make redis-cli # Acessar Redis diretamente
make clean # Reset completo + volumes
```

### Monitoramento

```bash
# Estatísticas Redis
make redis-stats
# Chaves ativas no Redis
docker exec rate_limiter_redis redis-cli keys "\*"
# Monitor operações em tempo real
docker exec rate_limiter_redis redis-cli monitor
```

## 📚 Aprendizados

- **Rate Limiting avançado** com algoritmo Token Bucket
- **Strategy Pattern** para arquitetura extensível
- **Redis Pipeline** para operações atômicas em sistemas concorrentes
- **Middleware customizado** para frameworks Go (Gin)
- **Graceful degradation** para alta disponibilidade
- **Docker multi-stage** para otimização de imagens
- **Configuração flexível** com Viper (.env + environment vars)
- **Testes de integração** com Redis real
- **IP extraction** considerando proxies e load balancers

<br>

**Desenvolvido com ❤️ em Go para aprendizado de rate limiting**
