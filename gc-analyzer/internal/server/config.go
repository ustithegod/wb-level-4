package server

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultAddr      = ":8080"
	defaultGCPercent = 100
)

type Config struct {
	Addr      string
	GCPercent int
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Addr:      defaultAddr,
		GCPercent: defaultGCPercent,
	}

	if addr := os.Getenv("ADDR"); addr != "" {
		cfg.Addr = addr
	}

	if raw := os.Getenv("GC_PERCENT"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse GC_PERCENT: %w", err)
		}
		cfg.GCPercent = value
	}

	return cfg, nil
}
