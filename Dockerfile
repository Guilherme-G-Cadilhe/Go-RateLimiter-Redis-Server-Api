# Stage 1: Builder - Compila a aplicação
FROM golang:1.23-alpine AS builder

# Instala git (necessário para algumas dependências Go)
RUN apk add --no-cache git

# Define diretório de trabalho
WORKDIR /app

# Copia arquivos de dependências primeiro (para cache do Docker)
COPY go.mod go.sum ./

# Baixa dependências (cacheable se go.mod/go.sum não mudaram)
RUN go mod download

# Copia código fonte
COPY . .

# Compila a aplicação
# CGO_ENABLED=0: compila estaticamente (sem dependências C)
# GOOS=linux: força compilação para Linux
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Stage 2: Runtime - Imagem final mínima
FROM alpine:latest

# Instala certificados SSL (para HTTPS se necessário)
RUN apk --no-cache add ca-certificates tzdata

# Cria usuário não-root por segurança
RUN adduser -D -s /bin/sh appuser

# Define diretório de trabalho
WORKDIR /app

# Copia binário compilado do stage anterior
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Muda ownership para usuário não-root
RUN chown -R appuser:appuser /app
USER appuser

# Expõe porta da aplicação
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]
