package main

import "flag"

type config struct {
	addr      string
	dir       string
	selfcheck bool
}

func parseConfig() config {
	cfg := config{}
	flag.StringVar(&cfg.addr, "addr", "127.0.0.1:8091", "listen address")
	flag.StringVar(&cfg.dir, "dir", "./data", "data directory")
	flag.BoolVar(&cfg.selfcheck, "selfcheck", false, "run self check and exit")
	flag.Parse()
	return cfg
}
