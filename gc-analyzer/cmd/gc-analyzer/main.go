package main

import (
	"log"
	"net/http"

	"github.com/ustithegod/wb-level-4/gc-analyzer/internal/server"
)

func main() {
	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	srv := server.New(cfg)
	log.Printf("starting gc-analyzer on %s with GC_PERCENT=%d", cfg.Addr, cfg.GCPercent)

	if err := http.ListenAndServe(cfg.Addr, srv.Handler()); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
