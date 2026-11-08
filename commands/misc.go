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
	RegisterCommand("ping", measureTime)
	RegisterCommand("security", securityReminder)
	RegisterCommand("checkCounts", checkCounts)
	RegisterCommand("checkBotCounts", botCheckCounts)
	RegisterCommand("randomiseStatus", randomiseStatus)
	RegisterCommand("removeAllMyRoles", RemoveAllMyRoles)
	RegisterCommand("listAllErrorsInBuffer", ListAllErrorsInBuffer)
	RegisterCommand("author", authorPromo)
	RegisterCommand("tomato", toggleTomato)
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

func toggleTomato(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	if m.GuildID == "" {
		s.ChannelMessageSend(m.ChannelID, "This command can only be used in a server.")
		return
	}
	if len(args) != 1 || (strings.ToLower(args[0]) != "enable" && strings.ToLower(args[0]) != "disable") {
		s.ChannelMessageSend(m.ChannelID, "Usage: tomato <enable|disable>")
		return
	}
	if s.State == nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to verify your server state right now.")
		return
	}

	permissions, err := userChannelPermissions(s, m.GuildID, m.Author.ID, m.ChannelID)
	if err != nil {
		utils.Errors.AddError(err.Error())
		s.ChannelMessageSend(m.ChannelID, "Unable to verify your server permissions right now.\n error: "+err.Error())
		return
	}
	moderatorPermissions := int64(
		discordgo.PermissionAdministrator |
			discordgo.PermissionModerateMembers |
			discordgo.PermissionVoiceMuteMembers |
			discordgo.PermissionKickMembers |
			discordgo.PermissionBanMembers,
	)
	if permissions&moderatorPermissions == 0 {
		s.ChannelMessageSend(m.ChannelID, "You need Mute Members, Kick Members, or Ban Members permission to use this command.")
		return
	}

	enabled := strings.ToLower(args[0]) == "enable"
	discord.SetTomatoEnabled(m.GuildID, enabled)
	if enabled {
		s.ChannelMessageSend(m.ChannelID, "Tomato timeout voting is now enabled for this server.")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Tomato timeout voting is now disabled for this server.")
}

func userChannelPermissions(s *discordgo.Session, guildID, userID, channelID string) (int64, error) {
	if s.State == nil {
		return 0, fmt.Errorf("discord state is not initialized")
	}
	if permissions, err := s.State.UserChannelPermissions(userID, channelID); err == nil {
		return permissions, nil
	}

	channel, err := s.Channel(channelID)
	if err != nil {
		return 0, fmt.Errorf("fetch channel for permission check: %w", err)
	}
	guild, err := s.Guild(guildID)
	if err != nil {
		return 0, fmt.Errorf("fetch guild for permission check: %w", err)
	}
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return 0, fmt.Errorf("fetch member for permission check: %w", err)
	}

	member.GuildID = guildID
	if err := s.State.GuildAdd(guild); err != nil {
		return 0, fmt.Errorf("cache guild for permission check: %w", err)
	}
	if err := s.State.ChannelAdd(channel); err != nil {
		return 0, fmt.Errorf("cache channel for permission check: %w", err)
	}
	if err := s.State.MemberAdd(member); err != nil {
		return 0, fmt.Errorf("cache member for permission check: %w", err)
	}

	return s.State.UserChannelPermissions(userID, channelID)
}
