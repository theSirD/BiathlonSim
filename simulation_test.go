package main

import (
	"math"
	"testing"
	"time"
)

func createTestConfig() *Config {
	startTime, _ := ParseTime("10:00:00.000")
	startDelta, _ := ParseDuration("00:01:00")
	return &Config{
		Laps:          1,
		LapLen:        1000,
		PenaltyLen:    100,
		FiringLines:   1,
		StartTime:     startTime,
		StartDelta:    startDelta,
		StartStr:      "10:00:00.000",
		StartDeltaStr: "00:01:00",
	}
}

func TestSimulation_SimpleFinishNoPenalty(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulation(cfg)

	events := []Event{
		{Timestamp: testTime(9, 0, 0, 0), ID: EventRegistered, CompetitorID: 1},
		{Timestamp: testTime(9, 1, 0, 0), ID: EventStartTimeSet, CompetitorID: 1, ScheduledStartTime: testTime(10, 0, 0, 0)},
		{Timestamp: testTime(9, 59, 50, 0), ID: EventOnStartLine, CompetitorID: 1},
		{Timestamp: testTime(10, 0, 0, 0), ID: EventStarted, CompetitorID: 1},
		{Timestamp: testTime(10, 5, 0, 0), ID: EventOnFiringRange, CompetitorID: 1, FiringRange: 1},
		{Timestamp: testTime(10, 5, 1, 0), ID: EventTargetHit, CompetitorID: 1, Target: 1},
		{Timestamp: testTime(10, 5, 2, 0), ID: EventTargetHit, CompetitorID: 1, Target: 2},
		{Timestamp: testTime(10, 5, 3, 0), ID: EventTargetHit, CompetitorID: 1, Target: 3},
		{Timestamp: testTime(10, 5, 4, 0), ID: EventTargetHit, CompetitorID: 1, Target: 4},
		{Timestamp: testTime(10, 5, 5, 0), ID: EventTargetHit, CompetitorID: 1, Target: 5},
		{Timestamp: testTime(10, 5, 10, 0), ID: EventLeftFiringRange, CompetitorID: 1},
		{Timestamp: testTime(10, 10, 0, 0), ID: EventEndedMainLap, CompetitorID: 1},
	}

	sim.Run(events)
	sim.FinalizeResults()

	c, exists := sim.Competitors[1]
	if !exists {
		t.Fatal("Competitor 1 not found in simulation")
	}

	if c.Status != StatusCompleted {
		t.Errorf("Competitor status: got %s, want %s", c.Status, StatusCompleted)
	}
	if c.TotalHits != 5 {
		t.Errorf("Competitor TotalHits: got %d, want 5", c.TotalHits)
	}
	if c.TotalShots != 5 {
		t.Errorf("Competitor TotalShots: got %d, want 5", c.TotalShots)
	}
	if c.TotalPenaltiesServed != 0 {
		t.Errorf("Competitor TotalPenaltiesServed: got %d, want 0", c.TotalPenaltiesServed)
	}

	expectedFinishTime := testTime(10, 10, 0, 0)
	if !c.FinishTime.Equal(expectedFinishTime) {
		t.Errorf("Competitor FinishTime: got %v, want %v", c.FinishTime, expectedFinishTime)
	}

	if len(c.LapsData) != 1 {
		t.Fatalf("Expected 1 lap data, got %d", len(c.LapsData))
	}
	lap1 := c.LapsData[0]
	expectedLapDuration := 10 * time.Minute
	if lap1.LapDuration != expectedLapDuration {
		t.Errorf("Lap 1 Duration: got %v, want %v", lap1.LapDuration, expectedLapDuration)
	}
	expectedSpeed := float64(cfg.LapLen) / expectedLapDuration.Seconds()
	if math.Abs(lap1.AverageSpeed-expectedSpeed) > 0.001 {
		t.Errorf("Lap 1 AverageSpeed: got %f, want %f", lap1.AverageSpeed, expectedSpeed)
	}
}

func TestSimulation_FinishWithPenalty(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulation(cfg)

	events := []Event{
		{Timestamp: testTime(9, 0, 0, 0), ID: EventRegistered, CompetitorID: 2},
		{Timestamp: testTime(9, 1, 0, 0), ID: EventStartTimeSet, CompetitorID: 2, ScheduledStartTime: testTime(10, 0, 0, 0)},
		{Timestamp: testTime(10, 0, 0, 0), ID: EventStarted, CompetitorID: 2},
		{Timestamp: testTime(10, 5, 0, 0), ID: EventOnFiringRange, CompetitorID: 2, FiringRange: 1},
		{Timestamp: testTime(10, 5, 1, 0), ID: EventTargetHit, CompetitorID: 2, Target: 1},
		{Timestamp: testTime(10, 5, 10, 0), ID: EventLeftFiringRange, CompetitorID: 2},
		{Timestamp: testTime(10, 5, 15, 0), ID: EventEnteredPenaltyLaps, CompetitorID: 2},
		{Timestamp: testTime(10, 5, 15, 0).Add(4 * 30 * time.Second), ID: EventLeftPenaltyLaps, CompetitorID: 2},
		{Timestamp: testTime(10, 5, 15, 0).Add(4*30*time.Second + 5*time.Minute), ID: EventEndedMainLap, CompetitorID: 2},
	}

	sim.Run(events)
	sim.FinalizeResults()

	c, exists := sim.Competitors[2]
	if !exists {
		t.Fatal("Competitor 2 not found")
	}
	if c.Status != StatusCompleted {
		t.Errorf("Status: got %s, want %s", c.Status, StatusCompleted)
	}
	if c.TotalHits != 1 {
		t.Errorf("TotalHits: got %d, want 1", c.TotalHits)
	}
	if c.TotalShots != 5 {
		t.Errorf("TotalShots: got %d, want 5", c.TotalShots)
	}
	if c.TotalPenaltiesServed != 4 {
		t.Errorf("TotalPenaltiesServed: got %d, want 4", c.TotalPenaltiesServed)
	}
	expectedLapDuration := (testTime(10, 5, 15, 0).Add(4*30*time.Second + 5*time.Minute)).Sub(testTime(10, 0, 0, 0))

	if len(c.LapsData) != 1 {
		t.Fatalf("Expected 1 lap data, got %d", len(c.LapsData))
	}
	lap1 := c.LapsData[0]
	if lap1.LapDuration != expectedLapDuration {
		t.Errorf("Lap 1 Duration: got %v, want %v", lap1.LapDuration, expectedLapDuration)
	}
}

func TestSimulation_NotStarted(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulation(cfg)
	events := []Event{
		{Timestamp: testTime(9, 0, 0, 0), ID: EventRegistered, CompetitorID: 3},
		{Timestamp: testTime(9, 1, 0, 0), ID: EventStartTimeSet, CompetitorID: 3, ScheduledStartTime: testTime(10, 0, 0, 0)},
	}
	sim.Run(events)
	sim.FinalizeResults()

	c, exists := sim.Competitors[3]
	if !exists {
		t.Fatal("Competitor 3 not found")
	}
	if c.Status != StatusNotStarted {
		t.Errorf("Status: got %s, want %s", c.Status, StatusNotStarted)
	}
}

func TestSimulation_NotFinished(t *testing.T) {
	cfg := createTestConfig()
	sim := NewSimulation(cfg)
	events := []Event{
		{Timestamp: testTime(9, 0, 0, 0), ID: EventRegistered, CompetitorID: 4},
		{Timestamp: testTime(9, 1, 0, 0), ID: EventStartTimeSet, CompetitorID: 4, ScheduledStartTime: testTime(10, 0, 0, 0)},
		{Timestamp: testTime(10, 0, 0, 0), ID: EventStarted, CompetitorID: 4},
		{Timestamp: testTime(10, 5, 0, 0), ID: EventCannotContinue, CompetitorID: 4, Comment: "Injured"},
	}
	sim.Run(events)
	sim.FinalizeResults()

	c, exists := sim.Competitors[4]
	if !exists {
		t.Fatal("Competitor 4 not found")
	}
	if c.Status != StatusNotFinished {
		t.Errorf("Status: got %s, want %s", c.Status, StatusNotFinished)
	}
	if c.DNFComment != "Injured" {
		t.Errorf("DNFComment: got '%s', want 'Injured'", c.DNFComment)
	}
}
