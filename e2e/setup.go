package e2e

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type testSuiteSession struct {
	d              *discordgo.Session
	testeeClientID string
	channelID      string
	guildID        string
}

var (
	errReplyTimeout                = errors.New("Timed out when waiting for reply.")
	errAwaitTesteeReadinessTimeout = errors.New("Timed out when waiting for tested bot to come online.")
)

func setupTestSuite() *testSuiteSession {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Cannot load .env file:", err.Error())
	}
	token := os.Getenv("E2E_BOT_TOKEN")
	if token == "" {
		panic("Cannot find E2E token. You need to set one up first through the env var E2E_BOT_TOKEN.")
	}
	groBotClientID := os.Getenv("E2E_GROCER_BOT_ID")
	if groBotClientID == "" {
		panic("Cannot find GroceryBot's client ID. You need to set one up first through the env var E2E_GROCER_BOT_ID.")
	}
	channelID := os.Getenv("E2E_CHANNEL_ID")
	if channelID == "" {
		panic("Missing E2E_CHANNEL_ID.")
	}
	guildID := os.Getenv("E2E_GUILD_ID")
	if guildID == "" {
		panic("Missing E2E_GUILD_ID.")
	}
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	d.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildPresences | discordgo.IntentsGuildMembers
	if err := d.Open(); err != nil {
		panic(err)
	}
	tss := &testSuiteSession{
		d:              d,
		testeeClientID: groBotClientID,
		channelID:      channelID,
		guildID:        guildID,
	}
	log.Printf("Waiting for testee %s to be ready.", tss.testeeClientID)
	tss.AwaitTesteeReadiness()
	log.Printf("Testee %s is now ready.", tss.testeeClientID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		<-sc
		tss.Cleanup()
	}()

	return tss
}

func (tss *testSuiteSession) AwaitTesteeReadiness() {
	readyChan := make(chan bool)
	removePresenceHandler := tss.d.AddHandler(func(s *discordgo.Session, p *discordgo.PresenceUpdate) {
		if p.User.ID == tss.testeeClientID && p.Status == discordgo.StatusOnline {
			readyChan <- true
		}
	})
	defer removePresenceHandler()
	removeGuildChunkHandler := tss.d.AddHandler(func(s *discordgo.Session, gc *discordgo.GuildMembersChunk) {
		for _, p := range gc.Presences {
			if p.User.ID == tss.testeeClientID && p.Status == discordgo.StatusOnline {
				log.Println(p)
				readyChan <- true
			}
		}
	})
	defer removeGuildChunkHandler()
	if err := tss.d.RequestGuildMembers(tss.guildID, "", 0, true); err != nil {
		panic(err)
	}
	select {
	case <-readyChan:
		return
	case <-time.After(10 * time.Second):
		panic(errAwaitTesteeReadinessTimeout.Error())
	}
}

func (tss *testSuiteSession) Cleanup() {
	tss.d.Close()
}

func (tss *testSuiteSession) SendAndAwaitReply(msg string) *discordgo.Message {
	_, err := tss.d.ChannelMessageSend(tss.channelID, msg)
	if err != nil {
		panic(err)
	}
	return tss.AwaitReply()
}

func (tss *testSuiteSession) AwaitReply() *discordgo.Message {
	replyChan := make(chan *discordgo.Message)
	remove := tss.d.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == tss.testeeClientID {
			replyChan <- m.Message
		}
	})
	defer remove()
	select {
	case res := <-replyChan:
		return res
	case <-time.After(3 * time.Second):
		panic(errReplyTimeout.Error())
	}
}
