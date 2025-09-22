package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
)

// Get Secrets from ENV Vars

type SecretsStruct struct {
	DatabaseDSN string `env:"database_dsn"`
}

var (
	Secrets     SecretsStruct
	readSecrets sync.Once
)

func LoadSecrets() {
	readSecrets.Do(func() {
		// Reads and parses secrets from env vars
		err := env.Parse(&Secrets)
		if err != nil {
			panic("Env Vars parse failed")
		}
	})
}

func GetSecrets() SecretsStruct {
	return Secrets
}
