package scouting

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v2"
)

//go:embed static/default_scouting_config.yaml
var defaultScoutConfigFile embed.FS

type ScoutingConfig struct {
	Actions  []Action  `yaml:"actions" json:"actions"`
	Outcomes []Outcome `yaml:"outcomes" json:"outcomes"`
}

var DefaultScoutingConfig ScoutingConfig

func init() {
	data, err := defaultScoutConfigFile.ReadFile("static/default_scouting_config.yaml")
	if err != nil {
		panic(fmt.Sprintf("reading default scout config: %v", err))
	}

	if err = yaml.Unmarshal(data, &DefaultScoutingConfig); err != nil {
		panic(fmt.Sprintf("parsing default scout config: %v", err))
	}
}

type Action struct {
	Name    string         `yaml:"name" json:"name"`
	Options []ActionOption `yaml:"options" json:"options"`
}

type ActionOption struct {
	Name string `yaml:"name" json:"name"`
}

type Outcome struct {
	Name               string   `yaml:"name" json:"name"`
	Points             uint     `yaml:"points" json:"points"`
	EndedInShot        bool     `yaml:"ended_in_shot" json:"ended_in_shot"`
	PossibleFreeThrows uint     `yaml:"possible_free_throws" json:"possible_free_throws"`
	StatisticTags      []string `yaml:"statistic_tags" json:"statistic_tags"`
}
