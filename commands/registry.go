package commands

import (
	"antiCommunitySpammer/config"
	"antiCommunitySpammer/discord"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(m *discordgo.MessageCreate, s *discordgo.Session, args []string)

var Registry = make(map[string]CommandHandler)

func init() {
	RegisterCommand("help", HandleHelp)
}

func RegisterCommand(name string, handler CommandHandler) {
	Registry[name] = handler
}

func OnMessageCreateWithPrefix(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := config.BotConfig.Prefix
	if m.Author.Bot || prefix == "" {
		return
	}
	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	commandLine := strings.TrimPrefix(m.Content, prefix)
	args := strings.Fields(commandLine)
	if len(args) == 0 {
		return
	}

	handleCommand(s, m, args)
}

func handleCommand(session *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	cmdName := strings.ToLower(args[0])

	handler, exists := Registry[cmdName]
	if !exists {
		discord.SendMessageToGuildChannel("❌ Unknown command: "+cmdName, session, message.ChannelID, false)
		return
	}
	handler(message, session, args[1:])
}

func HandleHelp(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	if len(Registry) == 0 {
		s.ChannelMessageSend(m.ChannelID, "⚠️ No commands are registered.")
		return
	}

	commandNames := make([]string, 0, len(Registry))
	for name := range Registry {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	var builder strings.Builder
	for _, cmd := range commandNames {
		builder.WriteString(fmt.Sprintf("`%s`\n", cmd))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Bot Commands",
		Description: "Here’s a list of all available commands:\n\n" + builder.String(),
		Color:       0x5865F2, // Discord blurple
		Footer: &discordgo.MessageEmbedFooter{
			Text: "cheese",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://maxisjoe.xyz/res/icons/protectorbot.png",
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
