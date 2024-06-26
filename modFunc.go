package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func banUser(s *discordgo.Session, guildId string, userID string, reason string, channelId string) {
	var modMentions string = ""
	for _, modID := range mods {
		mention := fmt.Sprintf("<@&%s>", modID)
		modMentions += mention + " "
	}
	banErr := s.GuildBanCreateWithReason(guildId, userID, reason, 0)
	if banErr != nil {
		fmt.Println("Error banning user:", banErr)
		return
	}
	message := fmt.Sprintf("%s - User <@%s> has been banned. Reason: %s", modMentions, userID, reason)
	_, err := s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending ban message:", err)
		return
	}
}

func muteUser(s *discordgo.Session, guildID string, userID string, channelId string, reason string) error {

	err := s.GuildMemberRoleAdd(guildID, userID, config.MuteRole)
	if err != nil {
		fmt.Println("Error muting user:", err)
		return err
	}

	message := fmt.Sprintf("User <@%s> has been muted. Reason: %s", userID, reason)
	_, err = s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending ban message:", err)
		return err
	}

	return nil
}
