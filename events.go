package main

import (
	"log"
	"regexp"
	"strconv"

	"golang.org/x/net/html"
)

var (
	// TODO: Consider not throwing away the extra time minutes.
	minuteRegexp = regexp.MustCompile(`^([0-9]+)'(?:\+[0-9]+')?$`)
	nameRegexp   = regexp.MustCompile(`^([a-zA-Z]+).*$`)
)

type kind uint

const (
	eventGoal kind = iota
	eventYellowCard
)

type team uint

const (
	teamHome team = iota
	teamAway
)

// minute is wrapped in a struct in case it becomes desirable to track
// minutes in extra time.
type minute struct {
	uint
}

type event struct {
	kind   kind
	minute minute
	player string
	team   team
}

// parseEvents parses all the events from a match report.
func parseEvents(doc *html.Node) []event {
	var events []event
	var fn func(*html.Node)

	fn = func(n *html.Node) {
		if isEvent(n) {
			// Each event has two data columns, the first is for the home
			// team, the second is for the away team.
			var tds []*html.Node
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "td" {
					tds = append(tds, c)
				}
			}

			var td *html.Node
			var e event
			switch len(tds) {
			case 1:
				td = tds[0]
				e.team = teamHome
			case 2:
				td = tds[1]
				e.team = teamAway
			}

			for c := td.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "p" {
					for c := c.FirstChild; c != nil; c = c.NextSibling {
						if c.Type == html.ElementNode && c.Data == "event-icon" {
							for c := c.FirstChild; c != nil; c = c.NextSibling {
								if c.Type == html.ElementNode && c.Data == "img" {
									alt := getAttr(c, "alt")
									switch alt {
									case "GOAL":
										e.kind = eventGoal
									case "INCIDENT":
										e.kind = eventYellowCard
									}
								}
							}
						}
						if c.Type == html.ElementNode && c.Data == "span" {
							if hasAttr(c, "class", "minute") {
								matches := minuteRegexp.FindStringSubmatch(c.FirstChild.Data)
								if len(matches) == 2 {
									i, _ := strconv.Atoi(matches[1])
									e.minute = minute{uint(i)}
								} else {
									log.Println("got", c.FirstChild.Data)
								}
							}
							if hasAttr(c, "class", "incidents") {
								matches := nameRegexp.FindStringSubmatch(c.FirstChild.Data)
								if len(matches) == 2 {
									e.player = matches[1]
								}
							}
						}
					}
				}

				if e.player != "" {
					events = append(events, e)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fn(c)
		}
	}
	fn(doc)

	return events
}

// getAttr returns the value of the first attribute that matches the
// provided key.
// If not matching key is found, the result will be the empty string.
func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// hasAttr returns true if the provided HTML node has the provided
// attribute.
func hasAttr(n *html.Node, key, value string) bool {
	for _, a := range n.Attr {
		if a.Key == key && a.Val == value {
			return true
		}
	}
	return false
}

// isEvent returns true if the provided HTML node is the table row
// representing an in-game event such as a goal.
func isEvent(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "tr" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "td" {
				return true
			}
		}
	}
	return false
}
