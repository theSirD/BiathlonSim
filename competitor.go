package main

import (
	"fmt"
	"strings"
	"time"
)

type CompetitorStatus string

const (
	StatusRegistered   CompetitorStatus = "Registered"
	StatusScheduled    CompetitorStatus = "Scheduled"
	StatusRacing       CompetitorStatus = "Racing"
	StatusOnRange      CompetitorStatus = "OnRange"
	StatusInPenalty    CompetitorStatus = "InPenalty"
	StatusCompleted    CompetitorStatus = "Completed"
	StatusNotStarted   CompetitorStatus = "NotStarted"
	StatusNotFinished  CompetitorStatus = "NotFinished"
	StatusDisqualified CompetitorStatus = "Disqualified"
)

type LapRecord struct {
	LapNumber        int
	StartTime        time.Time
	EndTime          time.Time
	ShootingData     []ShootingRecord
	PenaltyEntryTime time.Time
	PenaltyExitTime  time.Time
	PenaltiesServed  int

	LapDuration  time.Duration
	AverageSpeed float64
}

type ShootingRecord struct {
	RangeID           int
	EntryTime         time.Time
	ExitTime          time.Time
	Hits              int
	Shots             int
	PenaltiesIncurred int
}

type PenaltyData struct {
	TotalTime    time.Duration
	TotalLaps    int
	AverageSpeed float64
}

type Competitor struct {
	ID                 int
	Status             CompetitorStatus
	ScheduledStartTime time.Time
	ActualStartTime    time.Time
	FinishTime         time.Time
	LastEventTime      time.Time

	CurrentLapNumber   int
	LapsData           []LapRecord
	CurrentShooting    *ShootingRecord
	CurrentLapTempData struct {
		LapStartTime     time.Time
		RangeEntryTime   time.Time
		ShotsInSession   int
		HitsInSession    int
		PenaltiesToServe int
		PenaltyEntryTime time.Time
	}

	TotalHits            int
	TotalShots           int
	TotalPenaltiesServed int

	DNFComment             string
	DisqualificationReason string

	GeneratedEvents []Event
}

func NewCompetitor(id int) *Competitor {
	return &Competitor{
		ID:               id,
		Status:           StatusRegistered,
		LapsData:         make([]LapRecord, 0),
		CurrentLapNumber: 0,
	}
}

func GetOrCreateCompetitor(id int, competitors map[int]*Competitor) *Competitor {
	if c, ok := competitors[id]; ok {
		return c
	}
	c := NewCompetitor(id)
	competitors[id] = c
	return c
}

func (c *Competitor) CalculateResults(config *Config) {
	for i := range c.LapsData {
		lap := &c.LapsData[i]
		if !lap.StartTime.IsZero() && !lap.EndTime.IsZero() {
			lap.LapDuration = lap.EndTime.Sub(lap.StartTime)
			if lap.LapDuration > 0 {
				lap.AverageSpeed = float64(config.LapLen) / lap.LapDuration.Seconds()
			}
		}
	}
}

func (c *Competitor) FormatLapResults(config *Config) string {
	var lapStrings []string
	for i := 0; i < config.Laps; i++ {
		if i < len(c.LapsData) {
			lap := c.LapsData[i]
			if !lap.EndTime.IsZero() {
				lapStrings = append(lapStrings, fmt.Sprintf("{%s, %.3f}", FormatDuration(lap.LapDuration), lap.AverageSpeed))
			} else if !lap.StartTime.IsZero() && c.Status == StatusNotFinished {
				lapStrings = append(lapStrings, "{,}")
			} else {
				lapStrings = append(lapStrings, "{,}")
			}
		} else {
			lapStrings = append(lapStrings, "{,}")
		}
	}
	return "[" + strings.Join(lapStrings, ", ") + "]"
}

func (c *Competitor) CalculatePenaltyStats(config *Config) PenaltyData {
	var totalPenaltyDuration time.Duration
	totalPenaltyLapsRun := 0

	for _, lap := range c.LapsData {
		if !lap.PenaltyEntryTime.IsZero() && !lap.PenaltyExitTime.IsZero() {
			duration := lap.PenaltyExitTime.Sub(lap.PenaltyEntryTime)
			totalPenaltyDuration += duration
			totalPenaltyLapsRun += lap.PenaltiesServed
		}
	}

	var avgSpeed float64
	if totalPenaltyDuration > 0 && totalPenaltyLapsRun > 0 && config.PenaltyLen > 0 {
		totalPenaltyDistance := float64(totalPenaltyLapsRun * config.PenaltyLen)
		avgSpeed = totalPenaltyDistance / totalPenaltyDuration.Seconds()
	}

	return PenaltyData{
		TotalTime:    totalPenaltyDuration,
		TotalLaps:    totalPenaltyLapsRun,
		AverageSpeed: avgSpeed,
	}
}

func (c *Competitor) GetOverallStatusForReport() string {
	if c.Status == StatusDisqualified {
		return fmt.Sprintf("[%s]", "Disqualified")
	}
	if c.Status == StatusNotFinished {
		return fmt.Sprintf("[%s]", "NotFinished")
	}
	if c.Status == StatusNotStarted {
		return fmt.Sprintf("[%s]", "NotStarted")
	}
	if c.Status == StatusCompleted && !c.FinishTime.IsZero() {
		totalRaceTime := c.FinishTime.Sub(c.ActualStartTime)
		return FormatDuration(totalRaceTime)
	}
	return fmt.Sprintf("[%s]", c.Status)
}

func (c *Competitor) FinalShootingString() string {
	return fmt.Sprintf("%d/%d", c.TotalHits, c.TotalShots)
}
