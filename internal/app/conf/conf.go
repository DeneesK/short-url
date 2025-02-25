package conf

import (
	"cmp"
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
	DBDSN                 string
	MigrationsPath        string
	MemoryUsageLimitBytes uint64
	SecretKey             string
}

var cfg ServerConf
var limit float64

func init() {
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.Float64Var(&limit, "memlimit", 1, "memory usage limit in Gb")
	flag.StringVar(&cfg.Env, "env", "dev", "environment: dev or prod")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "filepath to store dump")
	flag.StringVar(&cfg.DBDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.MigrationsPath, "mp", "file://migrations", "path to migrations, exp.: file://migrations")
}

func MustLoad() *ServerConf {
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
	if filename, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = filename
	}
	if dbURL, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DBDSN = dbURL
	}
	if secret, ok := os.LookupEnv("SECRET_KEY"); ok {
		cfg.SecretKey = secret
	}
	cfg.SecretKey = cmp.Or(cfg.SecretKey, "secret_key")
	return &cfg
}
