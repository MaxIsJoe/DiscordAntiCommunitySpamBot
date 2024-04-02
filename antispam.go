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
			sendWarningMessage(s, m, count, trackedTimeStr)
		}
	}

	if len(botMessagesToDelete) > 0 {
		for _, msg := range botMessagesToDelete {
			fmt.Println(time.Since(msg.Timestamp))
			if time.Since(msg.Timestamp) > config.TrackDuration/2 {
				fmt.Println("Removing " + msg.ID + " from bot message tracking.")
				err = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
				if err != nil {
					fmt.Println("Error deleting message:", err)
				}
				botMessagesToDelete = botMessagesToDelete[1:]
			} else {
				fmt.Println("Message " + msg.ID + " still in bot message tracking. " + time.Since(msg.Timestamp).String())
			}
		}
	}
}

func sendWarningMessage(s *discordgo.Session, m *discordgo.MessageCreate, count int, trackedTimeStr string) {
	message := fmt.Sprintf("Hey %s, you've repeated this message %d times in the past %s. Please slow down a bit!", m.Author.Username, count, trackedTimeStr)
	err := sendMessageToGuildChannel(message, s, m.ChannelID, true)
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
	if m.Content == "" {
		fmt.Println("Detected an empty message. Skipping.. If this message appears a lot, make sure that Discord are not being total asses with the message intent permission on larger servers.")
		return
	}
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
