package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

type Config struct {
	Token                  string        `json:"token"`
	TrackDuration          time.Duration `json:"track_duration"`
	RepeatLimit            int           `json:"repeat_limit"`
	Prefix                 string        `json:"prefix"`
	MuteRole               string        `json:"MuteRole"`
	RoleToRemove           string        `json:"RoleToRemove"`
	UnitystationServerList string        `json:"UnitystationServerList"`
}

const (
	configFile = "config.ini"
)

var BotConfig Config

func LoadConfig(filePath string) (*Config, error) {
	shouldLoadConfig := flag.Bool("o", false, "Load config from etc/secrets/")
	if *shouldLoadConfig {
		filePath = "./etc/secrets/config.ini"
	}
	cfg, err := ini.Load(filePath)
	if err != nil && os.IsNotExist(err) == false {
		return nil, errors.Wrapf(err, "failed to load config file: %s", filePath)
	} else if os.IsNotExist(err) {
		fmt.Println("Config file not found, creating a new one...")
		return CreateDefaultConfig(filePath)
	}

	err = cfg.Section("bot").StrictMapTo(&BotConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to map config section: bot")
	}

	if BotConfig.Token == "" {
		return nil, errors.New("token is empty")
	}

	if BotConfig.MuteRole == "" {
		return nil, errors.New("mute role is empty")
	}

	if BotConfig.MuteRole == "" {
		return nil, errors.New("talking perms role is empty")
	}

	if BotConfig.UnitystationServerList == "" {
		fmt.Println("Server list missing. This is required for one of the commands to work.")
	}

	LoadStatuesFromKeys(cfg.Section("status").Keys())
	LoadModsFromIniKeys(cfg.Section("moderators").Keys())

	fmt.Println("Loaded configuration:")
	fmt.Printf(" Token: %s\n", censorToken(BotConfig.Token))
	fmt.Printf(" TrackDuration: %s\n", BotConfig.TrackDuration)
	fmt.Printf(" RepeatLimit: %d\n", BotConfig.RepeatLimit)
	fmt.Printf(" Prefix: %s\n", BotConfig.Prefix)

	return &BotConfig, nil
}

func CreateDefaultConfig(filePath string) (*Config, error) {
	cfg := ini.Empty()
	err := cfg.Append([]byte(fmt.Sprintf("[bot]\nToken=\nTrackDuration=%dm\nRepeatLimit=4\n", 2*time.Minute)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create default config file: %s", filePath)
	}

	if err := cfg.SaveTo(filePath); err != nil {
		return nil, errors.Wrapf(err, "failed to save default config file: %s", filePath)
	}
	panic("Config file created, please edit it and start the bot again.")
}

func censorToken(token string) string {
	halfLength := len(token) / 2
	censoredPart := strings.Repeat("█", halfLength)
	return fmt.Sprintf("%s%s", token[:halfLength], censoredPart)
}
