package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	TimeLayout         = "15:04:05.000"
	TimeLayoutNoMillis = "15:04:05"
)

var projectEpoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func ParseTime(timeStr string) (time.Time, error) {
	var t time.Time
	var err error

	if strings.Contains(timeStr, ".") {
		t, err = time.Parse(TimeLayout, timeStr)
	} else {
		t, err = time.Parse(TimeLayoutNoMillis, timeStr)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time string '%s': %w", timeStr, err)
	}
	return time.Date(projectEpoch.Year(), projectEpoch.Month(), projectEpoch.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC), nil
}

func ParseDuration(durationStr string) (time.Duration, error) {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid duration format '%s', expected HH:MM:SS", durationStr)
	}

	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours in duration '%s': %w", durationStr, err)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes in duration '%s': %w", durationStr, err)
	}
	sPart := strings.Split(parts[2], ".")
	sec, err := strconv.Atoi(sPart[0])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds in duration '%s': %w", durationStr, err)
	}

	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second, nil
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	totalMilliseconds := d.Milliseconds()
	hours := totalMilliseconds / (1000 * 60 * 60)
	totalMilliseconds %= (1000 * 60 * 60)
	minutes := totalMilliseconds / (1000 * 60)
	totalMilliseconds %= (1000 * 60)
	seconds := totalMilliseconds / 1000
	milliseconds := totalMilliseconds % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

func FormatTime(t time.Time) string {
	return t.Format(TimeLayout)
}
