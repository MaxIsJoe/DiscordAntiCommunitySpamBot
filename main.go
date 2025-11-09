package main

import (
	"fmt"
	"time"

	"antiCommunitySpammer/commands"
	"antiCommunitySpammer/config"
	"antiCommunitySpammer/discord"

	"github.com/bwmarrin/discordgo"
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

func main() {
	config, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	discord.TrackedMessages = *discord.NewMessageCounts(config.TrackDuration)

	dg.AddHandler(discord.OnMessageCreate)
	dg.AddHandler(commands.OnMessageCreateWithPrefix)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening websocket connection:", err)
		return
	} else {
		fmt.Printf("Bot is now running as %s.  Press CTRL-C to exit.", dg.State.User.Username)
	}

	discord.SetRandomStatus(dg)

	<-make(chan struct{})

	// Close the connection
	dg.Close()
}
