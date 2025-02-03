package conf

import (
	"flag"
	"log"
	"os"
	"strconv"
)

const gbyte = 1_000_000_000

type ServerConf struct {
	ServerAddr            string
	BaseURL               string
	Env                   string
	FileStoragePath       string
	MemoryUsageLimitBytes uint64
}

func MustLoad() *ServerConf {
	var limit float64
	var cfg ServerConf
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.Float64Var(&limit, "memlimit", 1, "memory usage limit in Gb")
	flag.StringVar(&cfg.Env, "env", "dev", "environment: dev or prod")
	flag.StringVar(&cfg.FileStoragePath, "f", "./file_storage.json", "environment: dev or prod")
	flag.Parse()

	cfg.MemoryUsageLimitBytes = uint64(limit * gbyte)
	if serverAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddr = serverAddr
	}

	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = baseURL
	}

	if memLimit, ok := os.LookupEnv("MEM_USAGE_LIMIT_GB"); ok {
		limit, err := strconv.ParseFloat(memLimit, 64)
		if err != nil {
			log.Fatalf("failed to parse MEM_USAGE_LIMIT_GB: %v", err)
		}
		cfg.MemoryUsageLimitBytes = uint64(limit * gbyte)
	}

	if env, ok := os.LookupEnv("ENV"); ok {
		cfg.Env = env
	}
	if filiname, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = filiname
	}
	return &cfg
}
