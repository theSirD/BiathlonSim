package main

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		wantHour int
		wantMin  int
		wantSec  int
		wantNs   int
		wantErr  bool
	}{
		{"ValidTimeWithMillis", "10:20:30.123", 10, 20, 30, 123000000, false},
		{"ValidTimeNoMillis", "10:20:30", 10, 20, 30, 0, false},
		{"InvalidFormat", "10-20-30", 0, 0, 0, 0, true},
		{"InvalidHour", "25:20:30.123", 0, 0, 0, 0, true},
		{"BoundaryMillis", "00:00:00.000", 0, 0, 0, 0, false},
		{"BoundaryMillis999", "23:59:59.999", 23, 59, 59, 999000000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.timeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				year, month, day := projectEpoch.Date()
				expected := time.Date(year, month, day, tt.wantHour, tt.wantMin, tt.wantSec, tt.wantNs, time.UTC)
				if !got.Equal(expected) {
					t.Errorf("ParseTime() got = %v, want %v", got, expected)
				}
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name         string
		durationStr  string
		wantDuration time.Duration
		wantErr      bool
	}{
		{"ValidDurationSimple", "00:00:30", 30 * time.Second, false},
		{"ValidDurationComplex", "01:10:05", 1*time.Hour + 10*time.Minute + 5*time.Second, false},
		{"InvalidFormat", "00:30", 0, true},
		{"InvalidNumber", "00:AA:30", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.durationStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantDuration {
				t.Errorf("ParseDuration() got = %v, want %v", got, tt.wantDuration)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		wantStr  string
	}{
		{"ZeroDuration", 0, "00:00:00.000"},
		{"SimpleSeconds", 30*time.Second + 123*time.Millisecond, "00:00:30.123"},
		{"ComplexDuration", 1*time.Hour + 2*time.Minute + 3*time.Second + 456*time.Millisecond, "01:02:03.456"},
		{"OnlyMillis", 5 * time.Millisecond, "00:00:00.005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotStr := FormatDuration(tt.duration); gotStr != tt.wantStr {
				t.Errorf("FormatDuration() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tm := time.Date(projectEpoch.Year(), projectEpoch.Month(), projectEpoch.Day(), 10, 20, 30, 123000000, time.UTC)
	expected := "10:20:30.123"
	if got := FormatTime(tm); got != expected {
		t.Errorf("FormatTime() = %v, want %v", got, expected)
	}

	tmNoMillis := time.Date(projectEpoch.Year(), projectEpoch.Month(), projectEpoch.Day(), 11, 22, 33, 0, time.UTC)
	expectedNoMillis := "11:22:33.000"
	if got := FormatTime(tmNoMillis); got != expectedNoMillis {
		t.Errorf("FormatTime() = %v, want %v", got, expectedNoMillis)
	}
}
