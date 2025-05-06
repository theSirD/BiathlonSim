package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	validConfigContent := `{
		"laps": 2,
		"lapLen": 3500,
		"penaltyLen": 150,
		"firingLines": 2,
		"start": "10:00:00.000",
		"startDelta": "00:01:30"
	}`
	validConfigPath := filepath.Join(tempDir, "valid_config.json")
	if err := os.WriteFile(validConfigPath, []byte(validConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	invalidJsonConfigContent := `{ "laps": 2, "lapLen": "not_a_number" }`
	invalidJsonConfigPath := filepath.Join(tempDir, "invalid_json_config.json")
	if err := os.WriteFile(invalidJsonConfigPath, []byte(invalidJsonConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	invalidTimeConfigContent := `{
		"laps": 1, "lapLen": 1000, "penaltyLen": 100, "firingLines": 1,
		"start": "invalid-time", "startDelta": "00:00:30"
	}`
	invalidTimeConfigPath := filepath.Join(tempDir, "invalid_time_config.json")
	if err := os.WriteFile(invalidTimeConfigPath, []byte(invalidTimeConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	tests := []struct {
		name       string
		filePath   string
		wantLaps   int
		wantLapLen int
		wantStartH int
		wantDeltaS time.Duration
		wantErr    bool
	}{
		{"ValidConfig", validConfigPath, 2, 3500, 10, 90 * time.Second, false},
		{"FileNotFound", filepath.Join(tempDir, "non_existent_config.json"), 0, 0, 0, 0, true},
		{"InvalidJSON", invalidJsonConfigPath, 0, 0, 0, 0, true},
		{"InvalidTimeFormat", invalidTimeConfigPath, 0, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, err := LoadConfig(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if gotCfg.Laps != tt.wantLaps {
					t.Errorf("LoadConfig() Laps = %v, want %v", gotCfg.Laps, tt.wantLaps)
				}
				if gotCfg.LapLen != tt.wantLapLen {
					t.Errorf("LoadConfig() LapLen = %v, want %v", gotCfg.LapLen, tt.wantLapLen)
				}
				year, month, day := projectEpoch.Date()
				expectedStartTime := time.Date(year, month, day, tt.wantStartH, 0, 0, 0, time.UTC)
				if !gotCfg.StartTime.Equal(expectedStartTime) {
					t.Errorf("LoadConfig() StartTime = %v, want hour %v (got %v)", gotCfg.StartTime, tt.wantStartH, expectedStartTime)
				}
				if gotCfg.StartDelta != tt.wantDeltaS {
					t.Errorf("LoadConfig() StartDelta = %v, want %v", gotCfg.StartDelta, tt.wantDeltaS)
				}
			}
		})
	}
}
