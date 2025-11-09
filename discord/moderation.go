package discord

import (
	"antiCommunitySpammer/config"
	"antiCommunitySpammer/utils"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func BanUser(s *discordgo.Session, guildId string, userID string, reason string, channelId string) {
	var modMentions string = ""
	for _, modID := range config.Mods {
		mention := fmt.Sprintf("<@&%s>", modID)
		modMentions += mention + " "
	}
	banErr := s.GuildBanCreateWithReason(guildId, userID, reason, 0)
	if banErr != nil {
		fmt.Println("Error banning user:", banErr)
		utils.Errors.AddError(banErr.Error())
		return
	}
	message := fmt.Sprintf("%s - User <@%s> has been banned. Reason: %s", modMentions, userID, reason)
	_, err := s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending ban message:", err)
		utils.Errors.AddError(err.Error())
		return
	}
}

func MuteUser(s *discordgo.Session, guildID string, userID string, channelId string, reason string) error {

	if config.BotConfig.MuteRole == "" || config.BotConfig.RoleToRemove == "" {
		RemoveAllRoles(s, guildID, userID)
		fmt.Println("Roles not configured properly or at all, going with the nuclear option for safety.")
		return nil
	}

	err := s.GuildMemberRoleRemove(guildID, userID, config.BotConfig.RoleToRemove)
	if err != nil {
		fmt.Println("Error muting user:", err)
		utils.Errors.AddError(err.Error())
		return err
	}
	err = s.GuildMemberRoleAdd(guildID, userID, config.BotConfig.MuteRole)
	if err != nil {
		fmt.Println("Error while attempting to label user as muted:", err)
		utils.Errors.AddError(err.Error())
	}

	message := fmt.Sprintf("User <@%s> has been muted. Reason: %s", userID, reason)
	_, err = s.ChannelMessageSend(channelId, message)
	if err != nil {
		fmt.Println("Error sending mute message:", err)
		utils.Errors.AddError(err.Error())
		return err
	}

	return nil
}

func RemoveAllRoles(s *discordgo.Session, guildID string, userID string) {

	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		fmt.Println("Error getting member:", err)
		utils.Errors.AddError(err.Error())
		return
	}
	for _, roleID := range member.Roles {
		err = s.GuildMemberRoleRemove(guildID, userID, roleID)
		if err != nil {
			fmt.Println("Error removing role:", err)
			utils.Errors.AddError(err.Error())
			return
		}
	}
}
