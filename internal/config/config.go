package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Rate Limiter
	RateLimitIPRPS          int           `mapstructure:"RATE_LIMIT_IP_RPS"`
	RateLimitIPBlockTime    time.Duration `mapstructure:"RATE_LIMIT_IP_BLOCK_TIME"`
	RateLimitTokenRPS       int           `mapstructure:"RATE_LIMIT_TOKEN_RPS"`
	RateLimitTokenBlockTime time.Duration `mapstructure:"RATE_LIMIT_TOKEN_BLOCK_TIME"`

	// Redis
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// Server
	ServerPort string `mapstructure:"SERVER_PORT"`
}

func LoadConfig() *Config {
	// Configurar Viper para ler .env
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Permitir override por variáveis de ambiente
	viper.AutomaticEnv()

	// Ler arquivo .env
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Erro ao ler .env: %v", err)
	}

	// Valores padrão
	viper.SetDefault("RATE_LIMIT_IP_RPS", 10)
	viper.SetDefault("RATE_LIMIT_IP_BLOCK_TIME", "300s")
	viper.SetDefault("RATE_LIMIT_TOKEN_RPS", 100)
	viper.SetDefault("RATE_LIMIT_TOKEN_BLOCK_TIME", "600s")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("SERVER_PORT", "8080")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Erro ao fazer unmarshal da config: %v", err)
	}

	return &config
}
