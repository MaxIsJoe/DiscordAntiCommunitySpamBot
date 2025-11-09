package commands

import (
	"antiCommunitySpammer/config"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ServerList struct {
	Servers []ServerInfo `json:"servers"`
}

type ServerInfo struct {
	Passworded      bool   `json:"Passworded"`
	ServerName      string `json:"ServerName"`
	ForkName        string `json:"ForkName"`
	BuildVersion    int64  `json:"BuildVersion"`
	CurrentMap      string `json:"CurrentMap"`
	GameMode        string `json:"GameMode"`
	IngameTime      string `json:"IngameTime"`
	RoundTime       string `json:"RoundTime"`
	PlayerCount     int    `json:"PlayerCount"`
	PlayerCountMax  int    `json:"PlayerCountMax"`
	ServerIP        string `json:"ServerIP"`
	ServerPort      int    `json:"ServerPort"`
	WinDownload     string `json:"WinDownload"`
	OSXDownload     string `json:"OSXDownload"`
	LinuxDownload   string `json:"LinuxDownload"`
	FPS             int    `json:"fps"`
	GoodFileVersion string `json:"GoodFileVersion"`
}

func init() {
	RegisterCommand("servers", HandleServersCommand)
}

// fetchUnityServers contacts the UnityStation API and returns a list of servers.
func fetchUnityServers(serverListUrl string) ([]ServerInfo, error) {
	url := serverListUrl
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var list ServerList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	return list.Servers, nil
}

func HandleServersCommand(m *discordgo.MessageCreate, s *discordgo.Session, args []string) {
	servers, err := fetchUnityServers(config.BotConfig.UnitystationServerList)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to fetch servers: %v", err))
		return
	}

	if len(servers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No servers found.")
		return
	}

	for _, srv := range servers {
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("🛰️ %s", srv.ServerName),
			Color: 0x003E53,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Fork", Value: srv.ForkName, Inline: true},
				{Name: "Players", Value: fmt.Sprintf("%d/%d", srv.PlayerCount, srv.PlayerCountMax), Inline: true},
				{Name: "Map", Value: srv.CurrentMap, Inline: false},
				{Name: "Mode", Value: srv.GameMode, Inline: true},
				{Name: "Build", Value: fmt.Sprintf("%d (%s)", srv.BuildVersion, srv.GoodFileVersion), Inline: true},
				{Name: "Ingame Time", Value: srv.IngameTime, Inline: true},
				{Name: "Downloads", Value: fmt.Sprintf("[Windows](%s) | [Linux](%s) | [Mac](%s)",
					srv.WinDownload, srv.LinuxDownload, srv.OSXDownload), Inline: false},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Server IP: %s:%d | FPS: %d", srv.ServerIP, srv.ServerPort, srv.FPS),
			},
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	}
}
