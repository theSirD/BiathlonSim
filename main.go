package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	configFile := flag.String("config", "config.json", "Path to the configuration file")
	eventsFile := flag.String("events", "events", "Path to the events file")
	flag.Parse()

	baseDir := "."
	exePath, err := os.Executable()
	if err == nil {
		baseDir = filepath.Dir(exePath)
	}

	absConfigFile := *configFile
	if !filepath.IsAbs(absConfigFile) {
		absConfigFile = filepath.Join(baseDir, *configFile)
	}

	absEventsFile := *eventsFile
	if !filepath.IsAbs(absEventsFile) {
		absEventsFile = filepath.Join(baseDir, *eventsFile)
	}

	cfg, err := LoadConfig(absConfigFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	fmt.Printf("Configuration loaded from %s: %+v\n\n", absConfigFile, cfg)

	incomingEvents, err := LoadEvents(absEventsFile)
	if err != nil {
		log.Fatalf("Error loading events: %v", err)
	}
	fmt.Printf("Loaded %d events from %s.\n\n", len(incomingEvents), absEventsFile)

	simulation := NewSimulation(cfg)

	simulation.Run(incomingEvents)
	simulation.FinalizeResults()

	GenerateOutputLog(simulation.OutputLog)
	GenerateFinalReport(simulation.Competitors, cfg)

	fmt.Println("\nBiathlonSim finished.")
}
