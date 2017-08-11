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

type payload struct {
	Content string `json:"content"`
}

type configuration struct {
	Teams     []string
	Webhook   string
	Whitelist bool
}

type api struct {
	Matches []struct {
		Team1 struct {
			TeamName string `json:"team_name"`
		} `json:"team1"`
		Link          string `json:"link"`
		StarttimeUnix string `json:"starttime_unix"`
		Team2         struct {
			TeamName string `json:"team_name"`
		} `json:"team2"`
	} `json:"matches"`
}

func getMatches(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func sendNotification(text string, webhook string) {
	payload := payload{text}
	blob, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(blob))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
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
func loadConfiguration(path string, config *configuration) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Configuration file not found.. exiting")
		usage()
	}
	decoder := json.NewDecoder(file)

	err = decoder.Decode(config)
	if err != nil {
		panic(err)
	}

	if config.Webhook == "" {
		fmt.Println("Webhook is not set.. exiting")
		usage()
	}
}

func main() {
	configFile := flag.String("config", "", "Configuration file")
	flag.Parse()
	if *configFile == "" {
		fmt.Println("Configuration file not found.. exiting")
		usage()
	}
	configuration := &configuration{}
	loadConfiguration(*configFile, configuration)

	api := new(api)
	getMatches("http://dailydota2.com/match-api", api)
	if len(api.Matches) != 0 {
		for _, match := range api.Matches {
			team1 := match.Team1.TeamName
			team2 := match.Team2.TeamName
			send := false

			if configuration.Whitelist {
				send = whitelisted(team1, team2, configuration.Teams)
			} else {
				send = true
			}

			// Prevent multiple notifications by checking time difference
			startTime, _ := strconv.Atoi(match.StarttimeUnix)
			delta := time.Now().Unix() - int64(startTime)
			if send && delta <= 60 {
				// Link points to stream but the stats page is nicer, so replace
				// with the stats URL instead
				url := strings.Replace(match.Link, "match", "stats", 1)
				notification := fmt.Sprintf("%s vs %s :: %s", team1, team2, url)
				sendNotification(notification, configuration.Webhook)
			}
		}
	}
}
