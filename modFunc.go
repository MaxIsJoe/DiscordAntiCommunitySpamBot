package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func banUser(s *discordgo.Session, guildId string, userID string, reason string, channelId string) {
	message := fmt.Sprintf("User <@%s> has been banned. Reason: %s", userID, reason)
	_, err := s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending ban message:", err)
		return
	}
	s.GuildBanCreateWithReason(guildId, userID, reason, 0)
}

func muteUser(s *discordgo.Session, member *discordgo.Member, guildID string, userID string, channelId string, reason string) error {

	s.GuildMemberMute(guildID, member.User.ID, true)

	message := fmt.Sprintf("User <@%s> has been muted. Reason: %s", userID, reason)
	_, err := s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending ban message:", err)
		return err
	}

	return nil
}
