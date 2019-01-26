package dota

import (
	"errors"
	"net/http"
	"strings"
	"time"

	opendota "github.com/jasonodonnell/go-opendota"
)

type Dota struct {
	client *opendota.Client
	teams  map[int]string
}

type Match struct {
	Radiant string
	Dire    string
	League  string
}

func New(teams []string, httpClient *http.Client) (*Dota, error) {
	var dota Dota
	dota.teams = make(map[int]string)
	dota.client = opendota.NewClient(httpClient)
	if err := dota.getTeams(teams); err != nil {
		return nil, err
	}

	return &dota, nil
}

func (d Dota) GetMatches() ([]Match, error) {
	var results []Match
	if len(d.teams) == 0 {
		return nil, errors.New("No teams whitelisted, aborting")
	}

	games, _, err := d.client.LiveService.Live()
	if err != nil {
		return nil, err
	}

	for _, game := range games {
		now := time.Now().Unix()
		if (now - game.ActivateTime) > 60 {
			continue
		}

		if _, ok := d.teams[game.RadiantTeamID]; !ok {
			if _, ok := d.teams[game.DireTeamID]; !ok {
				continue
			}
		}

		var match Match
		leagues, _, err := d.client.LeagueService.Leagues()
		if err != nil {
			return nil, err
		}

		for _, league := range leagues {
			if league.LeagueID == game.LeagueID {
				match.League = league.Name
				if match.League == "" {
					match.League = "Unknown League"
				}
				break
			}
		}

		match.Radiant = game.RadiantTeamName
		match.Dire = game.DireTeamName
		results = append(results, match)
	}
	return results, nil
}

func (d Dota) getTeams(names []string) error {
	teams, _, err := d.client.TeamService.Teams()
	if err != nil {
		return err
	}
	for _, team := range teams {
		if lookupTeam(team.Name, names) {
			d.teams[team.TeamID] = team.Name
		}
	}
	return nil
}

func lookupTeam(name string, whitelist []string) bool {
	for _, whitelisted := range whitelist {
		if strings.ToLower(whitelisted) == strings.ToLower(name) {
			return true
		}
	}
	return false
}
