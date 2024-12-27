package conf

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type ServerConf struct {
	ServerAddr       string
	BaseURL          string
	MemoryUsageLimit float64
	MemoryCheckType  string
}

func MustLoad() *ServerConf {
	var cfg ServerConf
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.Float64Var(&cfg.MemoryUsageLimit, "mem-limit", 100, "memory usage limit in percents 0 <= mem-limit >=")
	flag.StringVar(&cfg.MemoryCheckType, "mem-type", "ram", "memory type for checking: ram, disk, both")
	flag.Parse()
	if serverAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddr = serverAddr
	}

	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = baseURL
	}
	if memLimit, ok := os.LookupEnv("MEM_USAGE_LIMIT"); ok {
		limit, err := strconv.ParseFloat(memLimit, 64)
		if err != nil {
			log.Fatalf("failed to parse MEM_USAGE_LIMIT: %v", err)
		} else if limit > 100 || limit < 0 {
			log.Fatal("failed to parse MEM_USAGE_LIMIT: must be 0 => n <= 100")
		}
		cfg.MemoryUsageLimit = limit
	}
	if memType, ok := os.LookupEnv("MEM_CHECKING_TYPE"); ok {
		if memType == "ram" || memType == "disk" || memType == "both" {
			cfg.MemoryCheckType = memType
		} else {
			log.Fatalf("failed to parse MEM_CHECKING_TYPE must be ram, disk or both")
		}
	}
	return &cfg
}
