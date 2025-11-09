package commands

import (
	"antiCommunitySpammer/config"
	"antiCommunitySpammer/discord"
	"antiCommunitySpammer/utils"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	RegisterCommand("servers", measureTime)
	RegisterCommand("security", securityReminder)
	RegisterCommand("checkCounts", checkCounts)
	RegisterCommand("checkBotCounts", botCheckCounts)
	RegisterCommand("randomiseStatus", randomiseStatus)
	RegisterCommand("removeAllMyRoles", RemoveAllMyRoles)
	RegisterCommand("listAllErrorsInBuffer", ListAllErrorsInBuffer)
	RegisterCommand("author", authorPromo)
}

func measureTime(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	var msgTime = m.Timestamp
	latency := time.Since(msgTime).Milliseconds()
	discord.SendMessageToGuildChannel(fmt.Sprintf("Pong! %dms", latency), s, m.ChannelID, false)
}

func securityReminder(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	discord.SendMessageToGuildChannel("# Security Reminder! \n\n 1.Avoid downloading any files from untrusted sources and random discord users.\n 2.Never use your discord account to login to any third party website, and never give any bot/service the ability to join servers for you.\n 3. Be aware of fake discord websites, they are designed to trick you into giving your login credentials.\n 4. Use software like Bitwarden, and keep a unique **strong** password for every account you use on the internet.", s, m.ChannelID, false)
	discord.SendMessageToGuildChannel("# Educate yourself! \n\n Discord scammers use various different ways to trick users into giving away their accounts for malicious use. You can stay up to date with how discord scammers trick users by keeping up to date with this playlist:\n\n https://www.youtube.com/watch?v=zh-HccsdXDQ&list=PLEqYobHF0_Nk50vPzBKZFdcYHMzEyhuU3", s, m.ChannelID, false)
}

func checkCounts(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	counts := ""
	for key, tracker := range discord.TrackedMessages.Data {
		userID := strings.Split(key, "-")[0]
		counts += fmt.Sprintf("\nUser %s has %d counts", userID, tracker.Count)
	}
	discord.SendMessageToGuildChannel(counts, s, m.ChannelID, false)
}

func botCheckCounts(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	botCount := len(discord.BotMessagesToDelete)
	counts := fmt.Sprintf("\nBot has %d counts", botCount)
	if botCount != 0 {
		counts += fmt.Sprintf("\ntime until next message delete: " + time.Since(discord.BotMessagesToDelete[0].Timestamp).String() + "/" + (config.BotConfig.TrackDuration / 2).String())
	}
	discord.SendMessageToGuildChannel(counts, s, m.ChannelID, false)
}

func authorPromo(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	s.ChannelMessageSend(m.ChannelID, "This bot is created by MaxIsJoe for the Unitystation Project's discord server.\n\nSupport the Author: https://maxisjoe.xyz/maxfund\n\nCheck out Unitystation: https://www.unitystation.org/")
}

func randomiseStatus(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	discord.SetRandomStatus(s)
}

func RemoveAllMyRoles(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	discord.RemoveAllRoles(s, m.GuildID, m.Author.ID)
}

func ListAllErrorsInBuffer(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	errorsDetected := utils.Errors.GetErrors()
	errorsText := ""
	for _, err := range errorsDetected {
		errorsText += err.Timestamp.String() + " - " + err.Message + "\n"
	}
	if errorsText == "" {
		errorsText = "No errors detected"
	}
	discord.SendMessageToGuildChannel(errorsText, s, m.ChannelID, false)
}
