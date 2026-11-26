package config

import (
	"crypto/rsa"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v4"
)

type Config struct {
	RunAddr string `env:"RUN_ADDR,required"`
	DBUri   string `env:"DB_URI,required"`
	AuthCfg AuthConfig
}

func ParseConfig() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.RunAddr, "addr", "localhost:8080", "run address")
	flag.StringVar(&cfg.DBUri, "db", "", "database uri")

	if err := env.Parse(&cfg); err != nil {
		log.Printf("error occured parsing env variables: %v", err)
		return nil, err
	}

	return &cfg, nil
}

type AuthConfig struct {
	AccessTTL        time.Duration `env:"ACCESS_TTL" envDefault:"15min"`
	RefreshTTL       time.Duration `env:"REFRESH_TTL" envDefault:"720h"`
	PrivateKeyB64    string        `env:"JWT_PRIVATE_KEY,required"`
	PublicKeyB64     string        `env:"JWT_PUBLIC_KEY,required"`
	RefreshSecretB64 string        `env:"REFRESH_SECRET_B64,required"`
}

type KeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func (c *Config) LoadKeys() (*KeyPair, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(c.AuthCfg.PrivateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(c.AuthCfg.PublicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}
