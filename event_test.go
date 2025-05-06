package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testTime(h, m, s, ms int) time.Time {
	parsed, _ := ParseTime(FormatDuration(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(ms)*time.Millisecond))
	return parsed
}

func TestLoadEvents(t *testing.T) {
	tempDir := t.TempDir()

	validEventsContent := `
[09:05:59.867] 1 1
[09:15:00.841] 2 1 09:30:00.000
[09:49:31.659] 5 1 1
[09:49:33.123] 6 1 2
[09:59:05.321] 11 1 Lost in the forest
`
	validEventsPath := filepath.Join(tempDir, "valid_events.txt")
	if err := os.WriteFile(validEventsPath, []byte(strings.TrimSpace(validEventsContent)), 0644); err != nil {
		t.Fatalf("Failed to write temp events file: %v", err)
	}

	invalidEventsContent := `
[09:05:59.867] 1 1
this is not a valid event
[09:15:00.841] 2 1 09:30:00.000
`
	invalidEventsPath := filepath.Join(tempDir, "invalid_events.txt")
	if err := os.WriteFile(invalidEventsPath, []byte(strings.TrimSpace(invalidEventsContent)), 0644); err != nil {
		t.Fatalf("Failed to write temp events file: %v", err)
	}

	tests := []struct {
		name          string
		filePath      string
		wantNumEvents int
		wantErr       bool
		checkEvents   func(t *testing.T, events []Event)
	}{
		{
			name:          "ValidEvents",
			filePath:      validEventsPath,
			wantNumEvents: 5,
			wantErr:       false,
			checkEvents: func(t *testing.T, events []Event) {
				if len(events) != 5 {
					t.Fatalf("Expected 5 events, got %d", len(events))
				}
				ev0 := events[0]
				if ev0.ID != EventRegistered || ev0.CompetitorID != 1 || !ev0.Timestamp.Equal(testTime(9, 5, 59, 867)) {
					t.Errorf("Event 0 mismatch: got %+v", ev0)
				}
				ev1 := events[1]
				if ev1.ID != EventStartTimeSet || ev1.CompetitorID != 1 || !ev1.ScheduledStartTime.Equal(testTime(9, 30, 0, 0)) {
					t.Errorf("Event 1 (StartTimeSet) mismatch: got %+v, scheduled: %v", ev1, ev1.ScheduledStartTime)
				}
				ev3 := events[3]
				if ev3.ID != EventTargetHit || ev3.CompetitorID != 1 || ev3.Target != 2 {
					t.Errorf("Event 3 (TargetHit) mismatch: got %+v", ev3)
				}
				ev4 := events[4]
				if ev4.ID != EventCannotContinue || ev4.CompetitorID != 1 || ev4.Comment != "Lost in the forest" {
					t.Errorf("Event 4 (CannotContinue) mismatch: got %+v", ev4)
				}
			},
		},
		{
			name:          "FileNotFound",
			filePath:      filepath.Join(tempDir, "non_existent_events.txt"),
			wantNumEvents: 0,
			wantErr:       true,
		},
		{
			name:          "EventsWithInvalidLine",
			filePath:      invalidEventsPath,
			wantNumEvents: 2,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEvents, err := LoadEvents(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotEvents) != tt.wantNumEvents {
				t.Errorf("LoadEvents() got %d events, want %d", len(gotEvents), tt.wantNumEvents)
			}
			if tt.checkEvents != nil {
				tt.checkEvents(t, gotEvents)
			}
		})
	}
}

func TestGetEventDescription(t *testing.T) {
	regTime, _ := ParseTime("10:00:00.000")
	event := Event{Timestamp: regTime, ID: EventRegistered, CompetitorID: 101}
	desc := GetEventDescription(event)
	expectedDesc := "The competitor(101) registered"
	if desc != expectedDesc {
		t.Errorf("GetEventDescription() for EventRegistered: got '%s', want '%s'", desc, expectedDesc)
	}

	startTime, _ := ParseTime("10:05:00.000")
	scheduledTime, _ := ParseTime("10:30:00.000")
	event2 := Event{Timestamp: startTime, ID: EventStartTimeSet, CompetitorID: 102, ScheduledStartTime: scheduledTime}
	desc2 := GetEventDescription(event2)
	expectedDesc2 := "The start time for the competitor(102) was set by a draw to 10:30:00.000"
	if desc2 != expectedDesc2 {
		t.Errorf("GetEventDescription() for EventStartTimeSet: got '%s', want '%s'", desc2, expectedDesc2)
	}
}
