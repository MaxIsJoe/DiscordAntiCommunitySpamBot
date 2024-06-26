package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func onMessageCreateWithPrefix(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := config.Prefix
	if m.Author.Bot || prefix == "" {
		return
	}

	if strings.HasPrefix(m.Content, config.Prefix) {
		command := strings.TrimPrefix(m.Content, prefix)
		args := strings.Fields(command)
		handleCommand(s, m, args)
	}
}

func handleCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	command := args[0]
	switch command {
	case "ping":
		measureTime(m, s)
	case "hello":
		s.ChannelMessageSend(m.ChannelID, "Hello, "+m.Author.Mention()+"!")
	case "secrem":
		securityReminder(m, s)
	case "counts":
		checkCounts(&messageCounts, m, s)
	case "botcounts":
		botCheckCounts(m, s)
	case "author":
		authorPromo(s, m)
	case "randomstatus":
		randomiseStatus(s)
	case "removeallmyroles":
		RemoveAllMyRoles(s, m)
	case "help":
		sendMessageToGuildChannel("For commands, check https://github.com/MaxIsJoe/DiscordAntiCommunitySpamBot/blob/main/misc.go\n For contributing to Unitystation: https://unitystation.github.io/unitystation/contribution-guides/Development-Standards-Guide/", s, m.ChannelID, false)
	default:
		sendMessageToGuildChannel("Unknown command: "+command, s, m.ChannelID, false)
	}
}

func measureTime(m *discordgo.MessageCreate, s *discordgo.Session) {
	var msgTime = m.Timestamp
	latency := time.Since(msgTime).Milliseconds()
	sendMessageToGuildChannel(fmt.Sprintf("Pong! %dms", latency), s, m.ChannelID, false)
}

func securityReminder(m *discordgo.MessageCreate, s *discordgo.Session) {
	sendMessageToGuildChannel("# Security Reminder! \n\n 1.Avoid downloading any files from untrusted sources and random discord users.\n 2.Never use your discord account to login to any third party website, and never give any bot/service the ability to join servers for you.\n 3. Be aware of fake discord websites, they are designed to trick you into giving your login credentials.\n 4. Use software like Bitwarden, and keep a unique **strong** password for every account you use on the internet.", s, m.ChannelID, false)
	sendMessageToGuildChannel("# Educate yourself! \n\n Discord scammers use various different ways to trick users into giving away their accounts for malicious use. You can stay up to date with how discord scammers trick users by keeping up to date with this playlist:\n\n https://www.youtube.com/watch?v=zh-HccsdXDQ&list=PLEqYobHF0_Nk50vPzBKZFdcYHMzEyhuU3", s, m.ChannelID, false)
}

func checkCounts(mCounts *MessageCounts, m *discordgo.MessageCreate, s *discordgo.Session) {
	counts := ""
	for key, tracker := range mCounts.data {
		userID := strings.Split(key, "-")[0]
		counts += fmt.Sprintf("\nUser %s has %d counts", userID, tracker.count)
	}
	sendMessageToGuildChannel(counts, s, m.ChannelID, false)
}

func botCheckCounts(m *discordgo.MessageCreate, s *discordgo.Session) {
	botCount := len(botMessagesToDelete)
	counts := fmt.Sprintf("\nBot has %d counts", botCount)
	if botCount != 0 {
		counts += fmt.Sprintf("\ntime until next message delete: " + time.Since(botMessagesToDelete[0].Timestamp).String() + "/" + (config.TrackDuration / 2).String())
	}
	sendMessageToGuildChannel(counts, s, m.ChannelID, false)
}

func authorPromo(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "This bot is created by MaxIsJoe for the Unitystation Project's discord server.\n\nSupport the Author: https://maxisjoe.xyz/maxfund\n\nCheck out Unitystation: https://www.unitystation.org/")
}

func randomiseStatus(s *discordgo.Session) {
	setRandomStatus(s)
}

func RemoveAllMyRoles(s *discordgo.Session, m *discordgo.MessageCreate) {
	RemoveAllRoles(s, m.GuildID, m.Author.ID)
}
