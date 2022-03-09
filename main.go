package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"golang.org/x/net/html"
)

const (
	reportGlob    = "*-*-*.html"
	statsTemplate = "stats.html.tmpl"
)

// RawStats are statistics extracted from the match reports directly.
type RawStats struct {
	Games   int
	Goals   int
	Yellows int
}

// ComputedStats are computed from the RawStats.
type ComputedStats struct {
	GPG float64
}

type Stats struct {
	RawStats
	ComputedStats
}

type playerStats map[string]Stats

func main() {
	logger := log.New(os.Stderr, "", log.Llongfile)

	var pss = make(playerStats)

	files, _ := filepath.Glob(reportGlob)
	for _, f := range files {
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

		for _, player := range extractPlayers(doc) {
			ps := pss[player]
			ps.Games += 1
			pss[player] = ps
		}
		for _, event := range extractEvents(doc) {
			ps := pss[event.player]
			switch event.kind {
			case eventGoal:
				ps.Goals += 1
			case eventYellowCard:
				ps.Yellows += 1
			}
			pss[event.player] = ps
		}
	}

	// Add goals per game and remove any players that have not played a minimum number of games.
	for player, stats := range pss {
		stats.GPG = float64(stats.Goals) / float64(stats.Games)
		pss[player] = stats
		if stats.Games < 3 {
			delete(pss, player)
		}
	}

	st, err := ioutil.ReadFile(statsTemplate)
	if err != nil {
		logger.Fatalf("Could not read template file %s: %v", statsTemplate, err)
	}

	t := template.New("stats.html")
	t.Parse(string(st))

	if err := t.Execute(os.Stdout, pss); err != nil {
		logger.Fatalf("Unable to execute template: %v", err)
	}
}
