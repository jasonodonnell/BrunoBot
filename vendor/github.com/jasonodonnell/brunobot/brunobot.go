// Brunobot notifies a Discord channel when Professional Dota 2 games are live.
// This should be run via cron every minute to ensure only one notification.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/jasonodonnell/BrunoBot/discord"
	"github.com/jasonodonnell/brunobot/dota"
)

type configuration struct {
	Teams     []string
	Webhook   string
	Whitelist bool
}

func usage() {
	log.Fatal("Usage: brunobot --config=/path/to/config.json")
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
		log.Error("Configuration file not found. Exiting..")
		usage()
	}

	configuration, err := parseConfiguration(*configFile)
	if err != nil {
		log.Error("Error reading configuration file: %s", err)
		usage()
	}

	httpClient := &http.Client{}
	client, err := dota.New(configuration.Teams, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	matches, err := client.GetMatches()
	if err != nil {
		log.Fatal(err)
	}

	for _, match := range matches {
		format := "%s: %s vs %s :: Post Game Results - https://www.opendota.com/matches/%d"
		notification := fmt.Sprintf(format, match.League, match.Radiant, match.Dire, match.MatchID)
		if err := discord.Send(notification, configuration.Webhook); err != nil {
			log.Fatal(err)
		}
	}
}
