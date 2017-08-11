// Brunobot notifies a Discord channel when Professional Dota 2 games are live.
// This should be run via cron every minute to ensure only one notification.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const API_URL = "http://dailydota2.com/match-api"

type webhookPayload struct {
	Content string `json:"content"`
}

type configuration struct {
	Teams     []string
	Webhook   string
	Whitelist bool
}

type apiResponse struct {
	Matches   []match `json:"matches"`
	Timestamp int     `json:"timestamp"`
}

type match struct {
	Team1 struct {
		TeamName string `json:"team_name"`
	} `json:"team1"`
	Team2 struct {
		TeamName string `json:"team_name"`
	} `json:"team2"`
	Link          string `json:"link"`
	StarttimeUnix string `json:"starttime_unix"`
}

func getMatches(url string) (matches []match, err error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var r apiResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	if err != nil {
		return
	}

	matches = r.Matches

	return
}

func sendNotification(text, webhook string) (err error) {
	blob, err := json.Marshal(webhookPayload{text})
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(blob))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}

func usage() {
	fmt.Println("Usage: brunobot --config=/path/to/config.json")
	os.Exit(1)
}

func whitelisted(team1 string, team2 string, whitelist []string) bool {
	a := strings.ToLower(team1)
	b := strings.ToLower(team2)
	for _, team := range whitelist {
		team = strings.ToLower(team)
		if a == team || b == team {
			return true
		}
	}
	return false
}

// Webhook must be set in configuration
func parseConfiguration(path string) (config configuration, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return
	}

	if config.Webhook == "" {
		err = fmt.Errorf("Webhook is not set")
		return
	}

	return
}

func main() {
	configFile := flag.String("config", "", "Configuration file")
	flag.Parse()
	if *configFile == "" {
		fmt.Println("Configuration file not found.. exiting")
		usage()
	}

	// parse configuration
	configuration, err := parseConfiguration(*configFile)
	if err != nil {
		fmt.Println(err)
		usage()
	}

	// get current matches
	matches, err := getMatches(API_URL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check if any matches were returned
	if len(matches) == 0 {
		return
	}

	for _, match := range matches {
		team1 := match.Team1.TeamName
		team2 := match.Team2.TeamName

		// check if the team makes the cut
		if configuration.Whitelist && !whitelisted(team1, team2, configuration.Teams) {
			continue
		}

		// Prevent multiple notifications by checking time difference
		startTime, _ := strconv.Atoi(match.StarttimeUnix)
		delta := time.Now().Unix() - int64(startTime)
		if delta > 60 {
			continue
		}

		// Link points to stream but the stats page is nicer, so replace
		// with the stats URL instead
		url := strings.Replace(match.Link, "match", "stats", 1)
		notification := fmt.Sprintf("%s vs %s :: %s", team1, team2, url)
		err = sendNotification(notification, configuration.Webhook)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
