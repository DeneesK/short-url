package conf

import (
	"flag"
	"os"
)

type ServerConf struct {
	ServerAddr string `env:"SERVER_ADDRESS"`
	BaseURL    string `env:"BASE_URL"`
}

func MustLoad() *ServerConf {
	var cfg ServerConf
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.Parse()
	if serverAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddr = serverAddr
	}
	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = baseURL
	}
	return &cfg
}
