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
	TomatoEmoji           = "🍅"
	TomatoName            = "tomato"
	MaxTrackedMessages    = 20
	TomatoTimeoutCooldown = 30 * time.Minute
)

type TomatoTracker struct {
	reactions             map[string]int
	messageHistory        []string // Ordered list of messageIDs (oldest to newest)
	timedOutMessages      map[string]struct{}
	timeoutInProgress     map[string]struct{}
	userTimeouts          map[string]time.Time
	userTimeoutInProgress map[string]struct{}
	cooldownReminders     map[string]struct{}
	disabledGuilds        map[string]struct{}
	mu                    sync.RWMutex
}

var tomatoTracker = &TomatoTracker{
	reactions:             make(map[string]int),
	messageHistory:        make([]string, 0, MaxTrackedMessages),
	timedOutMessages:      make(map[string]struct{}),
	timeoutInProgress:     make(map[string]struct{}),
	userTimeouts:          make(map[string]time.Time),
	userTimeoutInProgress: make(map[string]struct{}),
	cooldownReminders:     make(map[string]struct{}),
	disabledGuilds:        make(map[string]struct{}),
}

func OnReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.Emoji.Name != TomatoName && r.Emoji.MessageFormat() != TomatoEmoji {
		return
	}

	channel, err := s.Channel(r.ChannelID)
	if err != nil {
		utils.Errors.AddError(err.Error())
		return
	}
	if !IsTomatoEnabled(channel.GuildID) {
		return
	}

	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
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
		now := time.Now()
		shouldTimeout, shouldRemind := tomatoTracker.beginTimeout(message.Author.ID, r.MessageID, now)
		if !shouldTimeout {
			if shouldRemind {
				reminderMsg := fmt.Sprintf(
					"User <@%s> cannot be timed out again yet because their 30-minute cooldown has not passed.",
					message.Author.ID,
				)
				if _, err := s.ChannelMessageSend(r.ChannelID, reminderMsg); err != nil {
					utils.Errors.AddError(err.Error())
				}
			}
			return
		}

		unmuteTime := time.Now().Add(timeoutDuration)

		err := s.GuildMemberTimeout(channel.GuildID, message.Author.ID, &unmuteTime)
		tomatoTracker.finishTimeout(message.Author.ID, r.MessageID, time.Now(), err == nil)
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

func SetTomatoEnabled(guildID string, enabled bool) {
	tomatoTracker.mu.Lock()
	defer tomatoTracker.mu.Unlock()

	if enabled {
		delete(tomatoTracker.disabledGuilds, guildID)
		return
	}
	tomatoTracker.disabledGuilds[guildID] = struct{}{}
}

func IsTomatoEnabled(guildID string) bool {
	tomatoTracker.mu.RLock()
	defer tomatoTracker.mu.RUnlock()
	_, disabled := tomatoTracker.disabledGuilds[guildID]
	return !disabled
}

func (t *TomatoTracker) beginTimeout(userID, messageID string, now time.Time) (bool, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.timedOutMessages[messageID]; exists {
		return false, false
	}
	if _, exists := t.timeoutInProgress[messageID]; exists {
		return false, false
	}
	if _, exists := t.userTimeoutInProgress[userID]; exists {
		return false, false
	}
	if lastTimeout, exists := t.userTimeouts[userID]; exists && now.Before(lastTimeout.Add(TomatoTimeoutCooldown)) {
		if _, reminded := t.cooldownReminders[userID]; !reminded {
			t.cooldownReminders[userID] = struct{}{}
			return false, true
		}
		return false, false
	}

	delete(t.cooldownReminders, userID)
	t.timeoutInProgress[messageID] = struct{}{}
	t.userTimeoutInProgress[userID] = struct{}{}
	return true, false
}

func (t *TomatoTracker) finishTimeout(userID, messageID string, now time.Time, succeeded bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.timeoutInProgress, messageID)
	delete(t.userTimeoutInProgress, userID)
	if succeeded {
		t.timedOutMessages[messageID] = struct{}{}
		t.userTimeouts[userID] = now
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
