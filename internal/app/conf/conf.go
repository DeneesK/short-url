package conf

import "flag"

type ServerConf struct {
	Addr     string
	BaseAddr string
}

func MustLoad() *ServerConf {
	var cfg ServerConf
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.Addr, "b", "http://localhost:8000/qsd54gFg", "base address of the resulting shortened URL")
	return &cfg
}
