package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type EventID int

const (
	EventRegistered         EventID = 1
	EventStartTimeSet       EventID = 2
	EventOnStartLine        EventID = 3
	EventStarted            EventID = 4
	EventOnFiringRange      EventID = 5
	EventTargetHit          EventID = 6
	EventLeftFiringRange    EventID = 7
	EventEnteredPenaltyLaps EventID = 8
	EventLeftPenaltyLaps    EventID = 9
	EventEndedMainLap       EventID = 10
	EventCannotContinue     EventID = 11

	EventDisqualified EventID = 32
	EventFinished     EventID = 33
)

type Event struct {
	Timestamp      time.Time
	ID             EventID
	CompetitorID   int
	ExtraParamsStr string
	Line           string

	ScheduledStartTime time.Time
	FiringRange        int
	Target             int
	Comment            string
}

var eventRegex = regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2}\.\d{3})\]\s+(\d+)\s+(\d+)(?:\s+(.*))?$`)
var sourcePrefixRegex = regexp.MustCompile(`^\\s*`)

func LoadEvents(filePath string) ([]Event, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open events file '%s': %w", filePath, err)
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		originalLine := line

		if line == "" {
			continue
		}

		line = sourcePrefixRegex.ReplaceAllString(line, "")

		matches := eventRegex.FindStringSubmatch(line)
		if matches == nil {
			fmt.Printf("Warning: Skipping malformed event line %d: %s\n", lineNumber, originalLine)
			continue
		}

		timestampStr := matches[1]
		eventIDStr := matches[2]
		competitorIDStr := matches[3]
		extraParamsStr := ""
		if len(matches) > 4 {
			extraParamsStr = strings.TrimSpace(matches[4])
		}

		timestamp, err := ParseTime(timestampStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse timestamp on line %d ('%s'): %v. Skipping event.\n", lineNumber, originalLine, err)
			continue
		}

		eventIDInt, err := strconv.Atoi(eventIDStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse EventID on line %d ('%s'): %v. Skipping event.\n", lineNumber, originalLine, err)
			continue
		}

		competitorID, err := strconv.Atoi(competitorIDStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse CompetitorID on line %d ('%s'): %v. Skipping event.\n", lineNumber, originalLine, err)
			continue
		}

		event := Event{
			Timestamp:      timestamp,
			ID:             EventID(eventIDInt),
			CompetitorID:   competitorID,
			ExtraParamsStr: extraParamsStr,
			Line:           originalLine,
		}

		switch event.ID {
		case EventStartTimeSet:
			event.ScheduledStartTime, err = ParseTime(extraParamsStr)
			if err != nil {
				fmt.Printf("Warning: Failed to parse ScheduledStartTime for event 2 on line %d ('%s'): %v. Skipping event.\n", lineNumber, originalLine, err)
				continue
			}
		case EventOnFiringRange:
			event.FiringRange, err = strconv.Atoi(extraParamsStr)
			if err != nil {
				fmt.Printf("Warning: Failed to parse FiringRange for event 5 on line %d ('%s'): %v. Skipping event.\n", lineNumber, originalLine, err)
				continue
			}
		case EventTargetHit:
			event.Target, err = strconv.Atoi(extraParamsStr)
			if err != nil {
				fmt.Printf("Warning: Failed to parse Target for event 6 on line %d ('%s'): %v. Skipping event.\n", lineNumber, originalLine, err)
				continue
			}
		case EventCannotContinue:
			event.Comment = extraParamsStr
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading events file '%s': %w", filePath, err)
	}

	return events, nil
}

func GetEventDescription(event Event) string {
	switch event.ID {
	case EventRegistered:
		return fmt.Sprintf("The competitor(%d) registered", event.CompetitorID)
	case EventStartTimeSet:
		return fmt.Sprintf("The start time for the competitor(%d) was set by a draw to %s", event.CompetitorID, FormatTime(event.ScheduledStartTime))
	case EventOnStartLine:
		return fmt.Sprintf("The competitor(%d) is on the start line", event.CompetitorID)
	case EventStarted:
		return fmt.Sprintf("The competitor(%d) has started", event.CompetitorID)
	case EventOnFiringRange:
		return fmt.Sprintf("The competitor(%d) is on the firing range(%d)", event.CompetitorID, event.FiringRange)
	case EventTargetHit:
		return fmt.Sprintf("The target(%d) has been hit by competitor(%d)", event.Target, event.CompetitorID)
	case EventLeftFiringRange:
		return fmt.Sprintf("The competitor(%d) left the firing range", event.CompetitorID)
	case EventEnteredPenaltyLaps:
		return fmt.Sprintf("The competitor(%d) entered the penalty laps", event.CompetitorID)
	case EventLeftPenaltyLaps:
		return fmt.Sprintf("The competitor(%d) left the penalty laps", event.CompetitorID)
	case EventEndedMainLap:
		return fmt.Sprintf("The competitor(%d) ended the main lap", event.CompetitorID)
	case EventCannotContinue:
		return fmt.Sprintf("The competitor(%d) can't continue: %s", event.CompetitorID, event.Comment)
	case EventDisqualified:
		return fmt.Sprintf("The competitor(%d) is disqualified", event.CompetitorID)
	case EventFinished:
		return fmt.Sprintf("The competitor(%d) has finished", event.CompetitorID)
	default:
		return fmt.Sprintf("Unknown event %d for competitor %d with params '%s'", event.ID, event.CompetitorID, event.ExtraParamsStr)
	}
}
