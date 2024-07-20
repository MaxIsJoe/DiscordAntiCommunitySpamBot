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
		errorBuffer.AddError(banErr.Error())
		return
	}
	message := fmt.Sprintf("%s - User <@%s> has been banned. Reason: %s", modMentions, userID, reason)
	_, err := s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending ban message:", err)
		errorBuffer.AddError(err.Error())
		return
	}
}

func muteUser(s *discordgo.Session, guildID string, userID string, channelId string, reason string) error {

	if config.MuteRole == "" || config.RoleToRemove == "" {
		RemoveAllRoles(s, guildID, userID)
		fmt.Println("Roles not configured properly or at all, going with the nuclear option for safety.")
		return nil
	}

	err := s.GuildMemberRoleRemove(guildID, userID, config.RoleToRemove)
	if err != nil {
		fmt.Println("Error muting user:", err)
		errorBuffer.AddError(err.Error())
		return err
	}
	err = s.GuildMemberRoleAdd(guildID, userID, config.MuteRole)
	if err != nil {
		fmt.Println("Error while attempting to label user as muted:", err)
		errorBuffer.AddError(err.Error())
	}

	message := fmt.Sprintf("User <@%s> has been muted. Reason: %s", userID, reason)
	_, err = s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending mute message:", err)
		errorBuffer.AddError(err.Error())
		return err
	}

	return nil
}

func RemoveAllRoles(s *discordgo.Session, guildID string, userID string) {

	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		fmt.Println("Error getting member:", err)
		errorBuffer.AddError(err.Error())
		return
	}
	for _, roleID := range member.Roles {
		err = s.GuildMemberRoleRemove(guildID, userID, roleID)
		if err != nil {
			fmt.Println("Error removing role:", err)
			errorBuffer.AddError(err.Error())
			return
		}
	}
}
