package commands

import (
	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)

var Registry = make(map[string]CommandHandler)

func RegisterCommand(name string, handler CommandHandler) {
	Registry[name] = handler
}
