package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

type Config struct {
	Token         string        `json:"token"`
	TrackDuration time.Duration `json:"track_duration"`
	RepeatLimit   int           `json:"repeat_limit"`
}

const (
	configFile = "config.ini"
)

var config Config
var statuses []string

func main() {

	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	messageCounts = *NewMessageCounts(config.TrackDuration)

	dg.AddHandler(onMessageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening websocket connection:", err)
		return
	} else {
		fmt.Printf("Bot is now running as %s.  Press CTRL-C to exit.", dg.State.User.Username)
	}

	setRandomStatus(dg)

	<-make(chan struct{})

	// Close the connection
	dg.Close()
}

func loadConfig(filePath string) (*Config, error) {
	cfg, err := ini.Load(filePath)
	if err != nil && os.IsNotExist(err) == false {
		return nil, errors.Wrapf(err, "failed to load config file: %s", filePath)
	} else if os.IsNotExist(err) {
		fmt.Println("Config file not found, creating a new one...")
		return createDefaultConfig(filePath)
	}

	err = cfg.Section("bot").StrictMapTo(&config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to map config section: bot")
	}

	if config.Token == "" {
		return nil, errors.New("token is empty")
	}

	for _, line := range cfg.Section("status").Keys() {
		statuses = append(statuses, line.Value())
		fmt.Printf("appending status: %s\n", line.Value())
	}

	for _, line := range cfg.Section("moderators").Keys() {
		mods = append(mods, line.Value())
		fmt.Printf("appending mod: %s\n", line.Value())
	}

	fmt.Println("Loaded configuration:")
	halfLength := len(config.Token) / 2
	censoredPart := strings.Repeat("█", halfLength)
	fmt.Printf(" Token: %s%s\n", config.Token[:halfLength], censoredPart)
	fmt.Printf(" TrackDuration: %s\n", config.TrackDuration)
	fmt.Printf(" RepeatLimit: %d\n", config.RepeatLimit)

	return &config, nil
}

func createDefaultConfig(filePath string) (*Config, error) {
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

func setRandomStatus(session *discordgo.Session) error {
	randomIndex := rand.Intn(len(statuses))
	randomStatus := statuses[randomIndex]
	return session.UpdateGameStatus(0, randomStatus)
}
