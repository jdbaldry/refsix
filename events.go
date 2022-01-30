package main

import (
	"log"
	"regexp"
	"strconv"

	"golang.org/x/net/html"
)

var (
	// TODO: Consider not throwing away the extra time minutes.
	minuteRegexp       = regexp.MustCompile(`^([0-9]+)'(?:\+[0-9]+')?$`)
	nameRegexp         = regexp.MustCompile(`^([a-zA-Z]+).*$`)
	teamListNameRegexp = regexp.MustCompile(`^ (?:Player|([a-zA-Z]+)).*$`)
)

type kind uint

const (
	eventUnknown kind = iota
	eventGoal
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

// extractPlayers extracts all players from a match report.
func extractPlayers(doc *html.Node) []string {
	var players []string

	var fn func(*html.Node)
	fn = func(n *html.Node) {
		if isPlayer(n) {
			// Each player table row has a single data cell with th
			// player number and name between
			// team, the second is for the away team.
			for c := n.FirstChild; c != nil; c = c.NextSibling {

				if c.Type == html.ElementNode && c.Data == "td" {
					for c := c.FirstChild; c != nil; c = c.NextSibling {
						if c.Type == html.ElementNode && c.Data == "span" && hasAttr(c, "codes", "codes") {
							for c := c.FirstChild; c != nil; c = c.NextSibling {
								if c.Type == html.TextNode {
									matches := teamListNameRegexp.FindStringSubmatch(c.Data)
									if len(matches) == 2 && matches[1] != "" {
										players = append(players, matches[1])
									}
								}
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fn(c)
		}
	}
	fn(doc)
	return players
}

// extractEvents parses all the events from a match report.
func extractEvents(doc *html.Node) []event {
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
									default:
										e.kind = eventUnknown
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

			}
			if e.player != "" && e.kind != eventUnknown {
				events = append(events, e)
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

// isPlayer returns true if the provided HTML node is the table row
// representing a player in a team list.
func isPlayer(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "tr" && hasAttr(n, "class", "player")
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
