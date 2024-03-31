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
	case "author":
		authorPromo(s, m)
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command: "+command)
	}
}

func measureTime(m *discordgo.MessageCreate, s *discordgo.Session) {
	var msgTime = m.Timestamp
	latency := time.Since(msgTime).Milliseconds()
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Pong! %dms", latency))
}

func securityReminder(m *discordgo.MessageCreate, s *discordgo.Session) {
	s.ChannelMessageSend(m.ChannelID, "# Security Reminder! \n\n 1.Avoid downloading any files from untrusted sources and random discord users.\n 2.Never use your discord account to login to any third party website, and never give any bot/service the ability to join servers for you.\n 3. Be aware of fake discord websites, they are designed to trick you into giving your login credentials.\n 4. Use software like Bitwarden, and keep a unique **strong** password for every account you use on the internet.")
	s.ChannelMessageSend(m.ChannelID, "# Educate yourself! \n\n Discord scammers use various different ways to trick users into giving away their accounts for malicious use. You can stay up to date with how discord scammers trick users by keeping up to date with this playlist:\n\n https://www.youtube.com/watch?v=zh-HccsdXDQ&list=PLEqYobHF0_Nk50vPzBKZFdcYHMzEyhuU3")
}

func checkCounts(mCounts *MessageCounts, m *discordgo.MessageCreate, s *discordgo.Session) {
	counts := ""
	for key, tracker := range mCounts.data {
		userID := strings.Split(key, "-")[0]
		counts += fmt.Sprintf("\nUser %s has %d counts", userID, tracker.count)
	}
	s.ChannelMessageSend(m.ChannelID, counts)
}

func authorPromo(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "This bot is created by MaxIsJoe for the Unitystation Project's discord server.\n\nSupport the Author: https://maxisjoe.xyz/maxfund\n\nCheck out Unitystation: https://www.unitystation.org/")
}
