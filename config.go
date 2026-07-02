package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type Event struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Color       string `json:"color"`
	Weight      int    `json:"weight"`
}

type Roulette struct {
	Name   string  `json:"-"` // имя из имени файла
	Events []Event `json:"events"`
}

func LoadRoulettes(configDir string) (map[string]*Roulette, error) {
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	roulettes := make(map[string]*Roulette)
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join(configDir, f.Name()))
		if err != nil {
			return nil, err
		}
		var events []Event
		if err := json.Unmarshal(data, &events); err != nil {
			return nil, err
		}
		name := f.Name()[:len(f.Name())-5] // убираем .json
		roulettes[name] = &Roulette{Name: name, Events: events}
	}
	return roulettes, nil
}
