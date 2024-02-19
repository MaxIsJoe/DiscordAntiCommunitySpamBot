package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type messageTracker struct {
	user      *discordgo.User
	message   string
	firstSeen time.Time
	count     int
}

type MessageCounts struct {
	mutex    sync.Mutex
	data     map[string]*messageTracker
	duration time.Duration
}

var messageCounts MessageCounts
var mods []string

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	trackMessage(&messageCounts, m)

	count, err := checkExceedingLimit(&messageCounts, m)
	if err != nil {
		fmt.Println("Error checking message limit:", err)
		return
	}

	if count > config.RepeatLimit {
		trackedTimeStr := config.TrackDuration.String()
		if count >= config.RepeatLimit+2 {
			muteUser(s, m.GuildID, m.Author.ID, m.ChannelID, "Spamming")
		}
		if isNewUser(m) {
			banUser(s, m.GuildID, m.Author.ID, "Suspicious new user is spamming the server.", m.ChannelID)
		} else {
			sendWarningMessage(s, m, count, mods, trackedTimeStr)
		}
	}
}

func sendWarningMessage(s *discordgo.Session, m *discordgo.MessageCreate, count int, mods []string, trackedTimeStr string) {
	var modMentions string
	for _, modID := range mods {
		mention := fmt.Sprintf("<@&%s>", modID)
		modMentions += mention + " "
	}

	message := fmt.Sprintf("%s -: Hey %s, you've repeated this message %d times in the past %s. Please slow down a bit!",
		modMentions, m.Author.Username, count, trackedTimeStr)
	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		fmt.Println("Error sending warning message:", err)
		return
	}
}

func NewMessageCounts(duration time.Duration) *MessageCounts {
	return &MessageCounts{
		mutex:    sync.Mutex{},
		data:     make(map[string]*messageTracker),
		duration: duration,
	}
}

func trackMessage(mCounts *MessageCounts, m *discordgo.MessageCreate) {
	mCounts.mutex.Lock()
	defer mCounts.mutex.Unlock()

	key := fmt.Sprintf("%s-%s", m.Author.ID, m.Content)
	tracker, ok := mCounts.data[key]

	if !ok {
		tracker = &messageTracker{
			user:      m.Author,
			message:   m.Content,
			firstSeen: time.Now(),
			count:     1,
		}
		mCounts.data[key] = tracker
	} else {
		tracker.count++
		tracker.firstSeen = time.Now()
	}

	time.AfterFunc(config.TrackDuration, func() {
		mCounts.mutex.Lock()
		defer mCounts.mutex.Unlock()

		key := fmt.Sprintf("%s-%s", m.Author.ID, m.Content)
		if time.Since(tracker.firstSeen) > config.TrackDuration {
			delete(mCounts.data, key)
		}
	})
}

func checkExceedingLimit(mCounts *MessageCounts, m *discordgo.MessageCreate) (int, error) {
	mCounts.mutex.Lock()
	defer mCounts.mutex.Unlock()

	key := fmt.Sprintf("%s-%s", m.Author.ID, m.Content)
	tracker, ok := mCounts.data[key]

	if !ok {
		return 0, nil
	}

	return tracker.count, nil
}

func isNewUser(message *discordgo.MessageCreate) bool {
	member := message.Member
	today := time.Now()
	joinedAt := member.JoinedAt
	timeSinceJoin := today.Sub(joinedAt)
	newbieThreshold := time.Hour
	isNew := timeSinceJoin <= newbieThreshold
	return isNew
}
