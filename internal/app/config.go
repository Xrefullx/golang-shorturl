package app

import "github.com/caarlos0/env/v6"

type ServerConfig struct {
	Port string `env:"SERVER_ADRESS"`
	Url  string `env:"URL"`
}

func EnviromentConfig(sc *ServerConfig) {
	_ = env.Parse(sc)
	if sc.Url == "" {
		sc.Url = "http://localhost:8080"
	}
	if sc.Port == "" {
		sc.Port = ":8080"
	}
}
