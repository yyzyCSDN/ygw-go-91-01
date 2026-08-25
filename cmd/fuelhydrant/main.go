package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := parseConfig()
	s := newService(cfg.dir)
	if cfg.selfcheck {
		s.runSelfCheck()
		return
	}
	server := &http.Server{Addr: cfg.addr, Handler: s.routes()}
	log.Printf("fuelhydrant listening on %s", cfg.addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
