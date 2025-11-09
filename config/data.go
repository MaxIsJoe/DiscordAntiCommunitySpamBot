package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

var Mods []string
var Statuses []string

func LoadModsFromIniKeys(keys []*ini.Key) {
	for _, key := range keys {
		Mods = append(Mods, key.Value())
		fmt.Printf("appending mod: %s\n", key.Value())
	}
}

func LoadStatuesFromKeys(keys []*ini.Key) {
	for _, key := range keys {
		Statuses = append(Statuses, key.Value())
		fmt.Printf("appending status: %s\n", key.Value())
	}
}
