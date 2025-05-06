package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Laps          int    `json:"laps"`
	LapLen        int    `json:"lapLen"`
	PenaltyLen    int    `json:"penaltyLen"`
	FiringLines   int    `json:"firingLines"`
	StartStr      string `json:"start"`
	StartDeltaStr string `json:"startDelta"`

	StartTime  time.Time
	StartDelta time.Duration
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config JSON from '%s': %w", filePath, err)
	}

	cfg.StartTime, err = ParseTime(cfg.StartStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config StartTime '%s': %w", cfg.StartStr, err)
	}

	cfg.StartDelta, err = ParseDuration(cfg.StartDeltaStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config StartDelta '%s': %w", cfg.StartDeltaStr, err)
	}

	return &cfg, nil
}
