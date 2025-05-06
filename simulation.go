package main

import (
	"fmt"
	"sort"
)

type Simulation struct {
	Config      *Config
	Competitors map[int]*Competitor
	OutputLog   []string
}

func NewSimulation(config *Config) *Simulation {
	return &Simulation{
		Config:      config,
		Competitors: make(map[int]*Competitor),
		OutputLog:   make([]string, 0),
	}
}

func (s *Simulation) Run(incomingEvents []Event) {
	sort.SliceStable(incomingEvents, func(i, j int) bool {
		return incomingEvents[i].Timestamp.Before(incomingEvents[j].Timestamp)
	})

	for _, event := range incomingEvents {
		s.processEvent(event)
	}

	s.checkForNotStarted()
}

func (s *Simulation) processEvent(event Event) {
	s.OutputLog = append(s.OutputLog, fmt.Sprintf("[%s] %s", FormatTime(event.Timestamp), GetEventDescription(event)))

	competitor := GetOrCreateCompetitor(event.CompetitorID, s.Competitors)
	competitor.LastEventTime = event.Timestamp

	switch event.ID {
	case EventRegistered:
		competitor.Status = StatusRegistered
	case EventStartTimeSet:
		competitor.ScheduledStartTime = event.ScheduledStartTime
		competitor.Status = StatusScheduled
	case EventOnStartLine:
	case EventStarted:
		if competitor.Status == StatusNotStarted || competitor.Status == StatusDisqualified {
			s.OutputLog = append(s.OutputLog, fmt.Sprintf("Warning: Competitor %d received Start event but is already %s.", competitor.ID, competitor.Status))
			return
		}
		competitor.ActualStartTime = event.Timestamp
		competitor.Status = StatusRacing
		competitor.CurrentLapNumber = 1
		competitor.CurrentLapTempData.LapStartTime = event.Timestamp
		if len(competitor.LapsData) == 0 {
			lapData := LapRecord{
				LapNumber:    competitor.CurrentLapNumber,
				StartTime:    event.Timestamp,
				ShootingData: make([]ShootingRecord, 0),
			}
			competitor.LapsData = append(competitor.LapsData, lapData)
		}

	case EventOnFiringRange:
		competitor.Status = StatusOnRange
		competitor.CurrentLapTempData.RangeEntryTime = event.Timestamp
		competitor.CurrentLapTempData.ShotsInSession = 0
		competitor.CurrentLapTempData.HitsInSession = 0
	case EventTargetHit:
		if competitor.Status != StatusOnRange {
			s.OutputLog = append(s.OutputLog, fmt.Sprintf("Warning: Competitor %d (%s) received TargetHit event but is not on firing range.", competitor.ID, competitor.Status))
		}
		competitor.CurrentLapTempData.HitsInSession++
		competitor.TotalHits++
	case EventLeftFiringRange:
		if competitor.Status != StatusOnRange {
			s.OutputLog = append(s.OutputLog, fmt.Sprintf("Warning: Competitor %d (%s) received LeftFiringRange event but was not on firing range.", competitor.ID, competitor.Status))
		}

		shotsThisSession := 5
		competitor.TotalShots += shotsThisSession

		penalties := shotsThisSession - competitor.CurrentLapTempData.HitsInSession
		competitor.CurrentLapTempData.PenaltiesToServe = penalties

		currentLapIdx := competitor.CurrentLapNumber - 1
		if currentLapIdx >= 0 && currentLapIdx < len(competitor.LapsData) {
			sr := ShootingRecord{
				RangeID:           event.FiringRange,
				EntryTime:         competitor.CurrentLapTempData.RangeEntryTime,
				ExitTime:          event.Timestamp,
				Hits:              competitor.CurrentLapTempData.HitsInSession,
				Shots:             shotsThisSession,
				PenaltiesIncurred: penalties,
			}
			competitor.LapsData[currentLapIdx].ShootingData = append(competitor.LapsData[currentLapIdx].ShootingData, sr)
		}

		if penalties > 0 {
			competitor.Status = StatusRacing
		} else {
			competitor.Status = StatusRacing
		}

	case EventEnteredPenaltyLaps:
		competitor.Status = StatusInPenalty
		competitor.CurrentLapTempData.PenaltyEntryTime = event.Timestamp
	case EventLeftPenaltyLaps:
		competitor.Status = StatusRacing

		lapIdx := competitor.CurrentLapNumber - 1
		if lapIdx >= 0 && lapIdx < len(competitor.LapsData) {
			competitor.LapsData[lapIdx].PenaltyEntryTime = competitor.CurrentLapTempData.PenaltyEntryTime
			competitor.LapsData[lapIdx].PenaltyExitTime = event.Timestamp
			competitor.LapsData[lapIdx].PenaltiesServed = competitor.CurrentLapTempData.PenaltiesToServe
			competitor.TotalPenaltiesServed += competitor.CurrentLapTempData.PenaltiesToServe
		}
		competitor.CurrentLapTempData.PenaltiesToServe = 0

	case EventEndedMainLap:
		lapIdx := competitor.CurrentLapNumber - 1
		if lapIdx >= 0 && lapIdx < len(competitor.LapsData) {
			competitor.LapsData[lapIdx].EndTime = event.Timestamp
		}

		if competitor.CurrentLapNumber == s.Config.Laps {
			competitor.Status = StatusCompleted
			competitor.FinishTime = event.Timestamp
			finishEvent := Event{
				Timestamp:    event.Timestamp,
				ID:           EventFinished,
				CompetitorID: competitor.ID,
			}
			competitor.GeneratedEvents = append(competitor.GeneratedEvents, finishEvent)
			s.OutputLog = append(s.OutputLog, fmt.Sprintf("[%s] %s", FormatTime(finishEvent.Timestamp), GetEventDescription(finishEvent)))
		} else {
			competitor.CurrentLapNumber++
			competitor.Status = StatusRacing
			competitor.CurrentLapTempData.LapStartTime = event.Timestamp
			if len(competitor.LapsData) < competitor.CurrentLapNumber {
				newLapData := LapRecord{
					LapNumber:    competitor.CurrentLapNumber,
					StartTime:    event.Timestamp,
					ShootingData: make([]ShootingRecord, 0),
				}
				competitor.LapsData = append(competitor.LapsData, newLapData)
			} else {
				competitor.LapsData[competitor.CurrentLapNumber-1].StartTime = event.Timestamp
			}
		}

	case EventCannotContinue:
		competitor.Status = StatusNotFinished
		competitor.DNFComment = event.Comment
	}
}

func (s *Simulation) checkForNotStarted() {

	for _, c := range s.Competitors {
		if c.ActualStartTime.IsZero() && (c.Status == StatusRegistered || c.Status == StatusScheduled) {
			c.Status = StatusNotStarted
		}

		if !c.ScheduledStartTime.IsZero() && c.ActualStartTime.IsZero() && c.Status != StatusDisqualified {
		}
	}
}

func (s *Simulation) FinalizeResults() {
	for _, competitor := range s.Competitors {
		competitor.CalculateResults(s.Config)

		if !competitor.ScheduledStartTime.IsZero() && competitor.ActualStartTime.IsZero() &&
			competitor.Status != StatusDisqualified && competitor.Status != StatusNotFinished {
			competitor.Status = StatusNotStarted
		}
		if competitor.ScheduledStartTime.IsZero() && competitor.ActualStartTime.IsZero() &&
			competitor.Status == StatusRegistered {
			competitor.Status = StatusNotStarted
		}
	}
}
