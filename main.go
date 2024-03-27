package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"stop-noti/util"
	"syscall"
	"time"
)

type Message struct {
	Content        string
	Time           time.Time
	Count          int
	MessageHistory []MessageHistory
}

type MessageHistory struct {
	ChannelID string
	MessageID string
	Delete    bool
}

var (
	dg  *discordgo.Session
	mdb = make(map[string]Message)
)

func init() {
	var err error
	config := util.GetConfig()
	dg, err = discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}
}

func main() {
	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err := dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	fmt.Printf("Bot(%s) is now running. Press CTRL-C to exit. \n", dg.State.User.ID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	err = dg.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("\nBot is now closed.")

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		fmt.Println(err)
		return
	}

	if m.Author.ID == guild.OwnerID {
		return
	}
	if util.GetConfig().WhiteList != nil {
		for _, v := range util.GetConfig().WhiteList {
			if m.Author.ID == v {
				return
			}
		}
	}

	if _, ok := mdb[m.Author.ID]; ok {
		if timeLapse(mdb[m.Author.ID].Time) && mdb[m.Author.ID].Content == m.Content {
			mdb[m.Author.ID] = Message{
				Content: mdb[m.Author.ID].Content,
				Time:    mdb[m.Author.ID].Time,
				Count:   mdb[m.Author.ID].Count + 1,
				MessageHistory: append(mdb[m.Author.ID].MessageHistory, MessageHistory{
					ChannelID: m.ChannelID,
					MessageID: m.ID,
					Delete:    false,
				}),
			}
			if mdb[m.Author.ID].Count == util.GetConfig().LimitCount {
				removeAllRolesFromUser(s, m)
			}
			if mdb[m.Author.ID].Count > util.GetConfig().LimitCount {
				removeMessage(s, m.ChannelID, m.ID)
			}
			return
		}
	}
	mdb[m.Author.ID] = Message{
		Content: m.Content,
		Time:    time.Now(),
		Count:   1,
		MessageHistory: []MessageHistory{
			{
				ChannelID: m.ChannelID,
				MessageID: m.ID,
				Delete:    false,
			},
		},
	}
	return
}

func timeLapse(t time.Time) bool {
	return time.Since(t) < (time.Duration(util.GetConfig().SummonTimeout) * time.Second)
}

func removeAllRolesFromUser(s *discordgo.Session, m *discordgo.MessageCreate) {
	userID := m.Author.ID
	serverID := m.GuildID

	roles, err := s.GuildMember(serverID, userID)
	if err != nil {
		fmt.Printf("Error getting roles for user %s in server %s: %s\n", userID, serverID, err)
		return
	}

	for _, roleID := range roles.Roles {
		err := s.GuildMemberRoleRemove(serverID, userID, roleID)
		if err != nil {
			fmt.Printf("Error removing role %s for user %s in server %s: %s\n", roleID, userID, serverID, err)
		}
	}
	fmt.Printf("All roles removed for user %s(%s) in server %s\n", userID, m.Author.Username, serverID)
	removeMessageHistory(s, m)
}

func removeMessageHistory(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, v := range mdb[m.Author.ID].MessageHistory {
		if v.Delete {
			continue
		}
		v.Delete = true
		removeMessage(s, v.ChannelID, v.MessageID)
	}
	delete(mdb, m.Author.ID)
	fmt.Printf("All messages deleted for user %s(%s)\n", m.Author.ID, m.Author.Username)
}

func removeMessage(s *discordgo.Session, channelID string, messageID string) {
	err := s.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		fmt.Printf("Error deleting message %s in channel %s: %s\n", messageID, channelID, err)
	}
}
