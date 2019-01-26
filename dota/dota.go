package dota

import (
	"errors"
	"net/http"
	"sort"
	"strings"

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
		if game.GameTime > 60 {
			continue
		}

		teams := make(map[string]int)
		for _, player := range game.Players {
			if _, ok := d.teams[player.TeamID]; ok {
				teams[player.TeamName]++
			}
		}
		pairs := sortMap(teams)

		if len(pairs) >= 2 && pairs[0].Value >= 5 {
			var match Match
			leagues, _, err := d.client.LeagueService.Leagues()
			if err != nil {
				return nil, err
			}

			for _, league := range leagues {
				if league.LeagueID == game.LeagueID {
					match.League = league.Name
				}
			}

			match.Radiant = pairs[0].Key
			match.Dire = pairs[1].Key
			results = append(results, match)
		}
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

func sortMap(teams map[string]int) PairList {
	pl := make(PairList, len(teams))
	i := 0
	for k, v := range teams {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
