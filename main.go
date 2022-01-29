package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/net/html"
)

const (
	reportGlob = "*-*-*.html"
)

type stats struct {
	games   int
	goals   int
	yellows int
}

type playerStats map[string]stats

func main() {
	logger := log.New(os.Stderr, "", log.Llongfile)

	var pss = make(playerStats)

	files, _ := filepath.Glob(reportGlob)
	for _, f := range files {
		logger.Printf("Processing file %q\n", f)
		absPath, err := filepath.Abs(f)
		if err != nil {
			logger.Fatalf("Could not determine absolute path to file: %v", err)
		}

		f, err := os.Open(absPath)
		if err != nil {
			logger.Fatalf("Could not open file %s: %v", absPath, err)
		}

		doc, err := html.Parse(f)
		if err != nil {
			logger.Fatalf("Could not parse file %s as HTML: %v", absPath, err)
		}

		for _, event := range parseEvents(doc) {
			ps := pss[event.player]
			switch event.kind {
			case eventGoal:
				ps.goals += 1
			case eventYellowCard:
				ps.yellows += 1
			}
			pss[event.player] = ps
		}
	}

	fmt.Printf("Name\tGoals\tYellows\n")
	for player, stats := range pss {
		fmt.Printf("%s\t%d\t%d\n", player, stats.goals/2, stats.yellows/2)
	}
}
