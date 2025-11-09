package discord

import (
	"antiCommunitySpammer/config"
	"antiCommunitySpammer/utils"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type messageTracker struct {
	user      *discordgo.User
	message   string
	firstSeen time.Time
	Count     int
}

type MessageCounts struct {
	mutex    sync.Mutex
	Data     map[string]*messageTracker
	Duration time.Duration
}

var TrackedMessages MessageCounts
var BotMessagesToDelete []discordgo.Message

func OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	trackMessage(&TrackedMessages, m)

	count, err := checkExceedingLimit(&TrackedMessages, m)
	if err != nil {
		fmt.Println("Error checking message limit:", err)
		utils.Errors.AddError(err.Error())
		return
	}

	if count > config.BotConfig.RepeatLimit {
		trackedTimeStr := config.BotConfig.TrackDuration.String()
		if count >= config.BotConfig.RepeatLimit+2 {
			MuteUser(s, m.GuildID, m.Author.ID, m.ChannelID, "Spamming")
		}
		if isNewUser(m) {
			BanUser(s, m.GuildID, m.Author.ID, "Suspicious new user is spamming the server.", m.ChannelID)
		} else {
			sendWarningMessage(s, m, count, trackedTimeStr)
		}
	}

	if len(BotMessagesToDelete) > 0 {
		for _, msg := range BotMessagesToDelete {
			fmt.Println(time.Since(msg.Timestamp))
			if time.Since(msg.Timestamp) > config.BotConfig.TrackDuration/2 {
				fmt.Println("Removing " + msg.ID + " from bot message tracking.")
				err = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
				if err != nil {
					fmt.Println("Error deleting message:", err)
					utils.Errors.AddError(err.Error())
				}
				BotMessagesToDelete = BotMessagesToDelete[1:]
			}
		}
	}
}

func sendWarningMessage(s *discordgo.Session, m *discordgo.MessageCreate, count int, trackedTimeStr string) {
	message := fmt.Sprintf("Hey %s, you've repeated this message %d times in the past %s. Please slow down a bit!", m.Author.Username, count, trackedTimeStr)
	err := SendMessageToGuildChannel(message, s, m.ChannelID, true)
	if err != nil {
		fmt.Println("Error sending warning message:", err)
		utils.Errors.AddError(err.Error())
		return
	}
}

func NewMessageCounts(duration time.Duration) *MessageCounts {
	return &MessageCounts{
		mutex:    sync.Mutex{},
		Data:     make(map[string]*messageTracker),
		Duration: duration,
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
	tracker, ok := mCounts.Data[key]

	if !ok {
		tracker = &messageTracker{
			user:      m.Author,
			message:   m.Content,
			firstSeen: time.Now(),
			Count:     1,
		}
		mCounts.Data[key] = tracker
	} else {
		tracker.Count++
		tracker.firstSeen = time.Now()
	}

	time.AfterFunc(config.BotConfig.TrackDuration, func() {
		mCounts.mutex.Lock()
		defer mCounts.mutex.Unlock()

		key := fmt.Sprintf("%s-%s", m.Author.ID, m.Content)
		if time.Since(tracker.firstSeen) > config.BotConfig.TrackDuration {
			delete(mCounts.Data, key)
		}
	})
}

func checkExceedingLimit(mCounts *MessageCounts, m *discordgo.MessageCreate) (int, error) {
	mCounts.mutex.Lock()
	defer mCounts.mutex.Unlock()

	key := fmt.Sprintf("%s-%s", m.Author.ID, m.Content)
	tracker, ok := mCounts.Data[key]

	if !ok {
		return 0, nil
	}

	return tracker.Count, nil
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

func SendMessageToGuildChannel(message string, session *discordgo.Session, channelID string, track bool) error {
	msg, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		fmt.Println("Error sending message - " + err.Error())
		return err
	}
	if track {
		BotMessagesToDelete = append(BotMessagesToDelete, *msg)
		fmt.Println("Added message to botMessagesToDelete - " + msg.Content + " - " + msg.Timestamp.String())
	}
	return nil
}
