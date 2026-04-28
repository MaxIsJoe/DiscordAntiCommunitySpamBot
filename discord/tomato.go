package discord

import (
	"antiCommunitySpammer/config"
	"antiCommunitySpammer/utils"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	TomatoEmoji        = "🍅"
	TomatoName         = "tomato"
	MaxTrackedMessages = 20
)

type TomatoTracker struct {
	reactions      map[string]int
	messageHistory []string // Ordered list of messageIDs (oldest to newest)
	mu             sync.RWMutex
}

var tomatoTracker = &TomatoTracker{
	reactions:      make(map[string]int),
	messageHistory: make([]string, 0, MaxTrackedMessages),
}

func OnReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.Emoji.Name != TomatoName && r.Emoji.MessageFormat() != TomatoEmoji {
		return
	}

	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		utils.Errors.AddError(err.Error())
		return
	}

	channel, err := s.Channel(r.ChannelID)
	if err != nil {
		utils.Errors.AddError(err.Error())
		return
	}

	// Track this message
	tomatoTracker.addMessage(r.MessageID)

	reactions := message.Reactions
	tomatoCount := 0
	for _, reaction := range reactions {
		if reaction.Emoji.Name == TomatoName || reaction.Emoji.MessageFormat() == TomatoEmoji {
			tomatoCount = reaction.Count
			break
		}
	}

	if tomatoCount >= config.BotConfig.TomatoReactionLimit {
		timeoutDuration := config.BotConfig.TomatoTimeoutDuration
		unmuteTime := time.Now().Add(timeoutDuration)

		err := s.GuildMemberTimeout(channel.GuildID, message.Author.ID, &unmuteTime)
		if err != nil {
			utils.Errors.AddError(err.Error())
			return
		}

		notificationMsg := fmt.Sprintf(
			"User <@%s> has been timed out for %v due to receiving %d tomato reactions on their message.",
			message.Author.ID,
			timeoutDuration,
			tomatoCount,
		)
		_, err = s.ChannelMessageSend(r.ChannelID, notificationMsg)
		if err != nil {
			utils.Errors.AddError(err.Error())
		}
	}
}

func OnReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if r.Emoji.Name != TomatoName && r.Emoji.MessageFormat() != TomatoEmoji {
		return
	}

	tomatoTracker.mu.Lock()
	defer tomatoTracker.mu.Unlock()

	if count, exists := tomatoTracker.reactions[r.MessageID]; exists && count > 0 {
		tomatoTracker.reactions[r.MessageID] = count - 1
	}
}

func (t *TomatoTracker) addMessage(messageID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If message already tracked, don't add it again.
	for _, id := range t.messageHistory {
		if id == messageID {
			return
		}
	}

	// Add new message
	t.messageHistory = append(t.messageHistory, messageID)
	t.reactions[messageID] = 0

	// Remove oldest message if we exceed the limit
	if len(t.messageHistory) > MaxTrackedMessages {
		oldest := t.messageHistory[0]
		t.messageHistory = t.messageHistory[1:]
		delete(t.reactions, oldest)
	}
}
