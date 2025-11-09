package discord

import (
	"antiCommunitySpammer/config"
	"math/rand"

	"github.com/bwmarrin/discordgo"
)

func SetRandomStatus(session *discordgo.Session) error {
	randomIndex := rand.Intn(len(config.Statuses))
	randomStatus := config.Statuses[randomIndex]
	return session.UpdateGameStatus(0, randomStatus)
}
