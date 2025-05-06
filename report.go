package main

import (
	"fmt"
	"sort"
)

func GenerateOutputLog(logEntries []string) {
	fmt.Println("Output log")
	fmt.Println("----------")
	for _, entry := range logEntries {
		fmt.Println(entry)
	}
	fmt.Println()
}

func GenerateFinalReport(competitors map[int]*Competitor, config *Config) {
	fmt.Println("Resulting table")
	fmt.Println("---------------")

	var sortedCompetitors []*Competitor
	for _, c := range competitors {
		sortedCompetitors = append(sortedCompetitors, c)
	}

	sort.Slice(sortedCompetitors, func(i, j int) bool {
		c1 := sortedCompetitors[i]
		c2 := sortedCompetitors[j]

		c1Finished := c1.Status == StatusCompleted && !c1.FinishTime.IsZero()
		c2Finished := c2.Status == StatusCompleted && !c2.FinishTime.IsZero()

		if c1Finished && c2Finished {
			t1 := c1.FinishTime.Sub(c1.ActualStartTime)
			t2 := c2.FinishTime.Sub(c2.ActualStartTime)
			if t1 != t2 {
				return t1 < t2
			}
			return c1.ID < c2.ID
		}
		if c1Finished {
			return true
		}
		if c2Finished {
			return false
		}

		statusOrder := func(s CompetitorStatus) int {
			switch s {
			case StatusNotFinished:
				return 1
			case StatusNotStarted:
				return 2
			case StatusDisqualified:
				return 3
			default:
				return 4
			}
		}
		if statusOrder(c1.Status) != statusOrder(c2.Status) {
			return statusOrder(c1.Status) < statusOrder(c2.Status)
		}
		return c1.ID < c2.ID
	})

	headerFormat := "%-15s %-5s %-45s %-23s %-10s\n"
	fmt.Printf(headerFormat, "Result/Status", "ID", "Lap Details (Time, Speed m/s)", "Penalty (Time, Speed m/s)", "Shooting")

	for _, c := range sortedCompetitors {
		statusStr := c.GetOverallStatusForReport()
		lapResultsStr := c.FormatLapResults(config)
		penaltyStats := c.CalculatePenaltyStats(config)
		penaltyStr := fmt.Sprintf("{%s, %.3f}", FormatDuration(penaltyStats.TotalTime), penaltyStats.AverageSpeed)
		if penaltyStats.TotalLaps == 0 {
			if penaltyStats.TotalTime == 0 {
				penaltyStr = fmt.Sprintf("{%s, 0.000}", FormatDuration(0))
			}
		}
		shootingStr := c.FinalShootingString()

		fmt.Printf("%-15s %-5d %-45s %-23s %-10s\n",
			statusStr,
			c.ID,
			lapResultsStr,
			penaltyStr,
			shootingStr)
	}
}
