package main

import (
	"encoding/json"
	"os"
	"sync"
)

const filePath = "config.json"

// FieldDef describes how to decode one named field from the raw IO output bytes.
type FieldDef struct {
	Name   string `json:"name"`
	Type   string `json:"type"`    // sint, int, dint, float32, real, bool
	Offset int    `json:"offset"`  // byte offset into the output buffer
	BitPos int    `json:"bit_pos"` // bit position within the byte (0-7), only used for bool
}

// Config is the full application configuration.
type Config struct {
	Fields               []FieldDef `json:"fields"`
	MaxFileSizeBytes     int64      `json:"max_file_size_bytes"`
	OutputDir            string     `json:"output_dir"`
	RotateIntervalHours  float64    `json:"rotate_interval_hours"`  // 0 = disabled
	MaxTotalSizeBytes    int64      `json:"max_total_size_bytes"`   // 0 = disabled
	MaxFileCount         int        `json:"max_file_count"`         // 0 = disabled
	FlushIntervalSeconds float64    `json:"flush_interval_seconds"` // 0 = disabled
}

var (
	mu      sync.RWMutex
	current Config
	// ChangeCh is signaled (non-blocking) every time the config is saved.
	// The historian selects on this to rotate to a new parquet file immediately.
	ChangeCh = make(chan struct{}, 1)
	// RotateCh is signaled by the web UI "Rotate Now" button.
	RotateCh = make(chan struct{}, 1)
	// ShutdownCh is signaled when the OS sends SIGINT or SIGTERM.
	ShutdownCh = make(chan struct{}, 1)
)

func LoadConfig() error {
	mu.Lock()
	defer mu.Unlock()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			current = defaults()
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &current)
}

func SaveConfig(c Config) error {
	mu.Lock()
	defer mu.Unlock()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}
	current = c
	select {
	case ChangeCh <- struct{}{}:
	default:
	}
	return nil
}

func GetConfig() Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func defaults() Config {
	return Config{
		MaxFileSizeBytes:     100 * 1024 * 1024, // 100 MB
		OutputDir:            "data",
		RotateIntervalHours:  24,
		MaxTotalSizeBytes:    10 * 1024 * 1024 * 1024, // 10 GB
		MaxFileCount:         365,
		FlushIntervalSeconds: 60,
		Fields:               []FieldDef{},
	}
}
