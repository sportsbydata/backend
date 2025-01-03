package server

import "github.com/sportsbydata/backend/scouting"

type scoutingConfig struct {
	Actions  []action  `json:"actions"`
	Outcomes []outcome `json:"outcomes"`
	Layouts  []layout  `json:"layouts"`
}

type layout struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

type action struct {
	ID      string         `json:"id"`
	Options []actionOption `json:"options"`
}

type actionOption struct {
	ID string `json:"id"`
}

type outcome struct {
	ID                 string   `json:"id"`
	Points             uint     `json:"points"`
	EndedInShot        bool     `json:"ended_in_shot"`
	PossibleFreeThrows uint     `json:"possible_free_throws"`
	StatisticTags      []string `json:"statistic_tags,omitempty"`
}

func newScoutingConfig(cfg scouting.ScoutingConfig) scoutingConfig {
	aa := make([]action, len(cfg.Actions))

	for i, a := range cfg.Actions {
		aa[i] = newAction(a)
	}

	oo := make([]outcome, len(cfg.Outcomes))

	for i, o := range cfg.Outcomes {
		oo[i] = newOutcome(o)
	}

	ll := make([]layout, len(cfg.Layouts))

	for i, l := range cfg.Layouts {
		ll[i] = newLayout(l)
	}

	return scoutingConfig{
		Actions:  aa,
		Outcomes: oo,
		Layouts:  ll,
	}
}

func newLayout(l scouting.Layout) layout {
	return layout{
		Name:    l.Name,
		Actions: l.Actions,
	}
}

func newActionOption(ao scouting.ActionOption) actionOption {
	return actionOption{
		ID: ao.ID,
	}
}

func newOutcome(o scouting.Outcome) outcome {
	return outcome{
		ID:                 o.ID,
		Points:             o.Points,
		EndedInShot:        o.EndedInShot,
		PossibleFreeThrows: o.PossibleFreeThrows,
		StatisticTags:      o.StatisticTags,
	}
}

func newAction(a scouting.Action) action {
	oo := make([]actionOption, len(a.Options))

	for i, o := range a.Options {
		oo[i] = newActionOption(o)
	}

	return action{
		ID:      a.ID,
		Options: oo,
	}
}
