package conf

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type ServerConf struct {
	ServerAddr string
	BaseURL    string
	RateLimit  int
}

func MustLoad() *ServerConf {
	var cfg ServerConf
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.IntVar(&cfg.RateLimit, "rate", 30, "request rate limit in minute")
	flag.Parse()
	if serverAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddr = serverAddr
	}
	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = baseURL
	}
	if limit, ok := os.LookupEnv("RATE_LIMIT"); ok {
		limit, err := strconv.Atoi(limit)
		if err != nil {
			log.Fatal("failed to convert RATE_LIMIT to int")
		} else {
			cfg.RateLimit = limit
		}
	}
	return &cfg
}
