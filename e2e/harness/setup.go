//go:build integration || healthcheck

package harness

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type TestSuiteSession struct {
	d              *discordgo.Session
	testeeClientID string
	channelID      string
	guildID        string
}

var (
	errReplyTimeout                = errors.New("timed out when waiting for reply")
	errAwaitTesteeReadinessTimeout = errors.New("timed out when waiting for tested bot to come online")
)

func dotEnvPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("harness.dotEnvPath: runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".env"))
}

func SetupTestSuite() *TestSuiteSession {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load(dotEnvPath()); err != nil {
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
	tss := &TestSuiteSession{
		d:              d,
		testeeClientID: groBotClientID,
		channelID:      channelID,
		guildID:        guildID,
	}
	if err := d.Open(); err != nil {
		panic(err)
	}
	defer tss.RecoverFromPanic()
	log.Printf("Waiting for testee %s to be ready.", tss.testeeClientID)
	tss.AwaitTesteeReadiness()
	log.Printf("Testee %s is now ready.", tss.testeeClientID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sc
		tss.Cleanup()
	}()

	return tss
}

func (tss *TestSuiteSession) RecoverFromPanic() {
	if r := recover(); r != nil {
		log.Println("Detected panic. Cleaning up session before panicking further.")
		tss.Cleanup()
		panic(r)
	}
}

func (tss *TestSuiteSession) AwaitTesteeReadiness() {
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
				readyChan <- true
			}
		}
	})
	defer removeGuildChunkHandler()
	if err := tss.d.RequestGuildMembers(tss.guildID, "", 0, "e2e_test", true); err != nil {
		panic(err)
	}
	select {
	case <-readyChan:
		return
	case <-time.After(10 * time.Second):
		panic(errAwaitTesteeReadinessTimeout.Error())
	}
}

func (tss *TestSuiteSession) Cleanup() {
	log.Println("Cleaning up test session...")
	tss.d.Close()
}

// ClientUserID returns the Discord user ID of the E2E test bot session.
func (tss *TestSuiteSession) ClientUserID() string {
	return tss.d.State.User.ID
}

func (tss *TestSuiteSession) SendAndAwaitReply(msg string) *discordgo.Message {
	// test that GroBot should work when mentioned (plus prod only accepts message commands if the bot is mentioned, so this is useful for the healthcheck)
	_, err := tss.d.ChannelMessageSend(tss.channelID, fmt.Sprintf("<@%s> %s", tss.testeeClientID, msg))
	if err != nil {
		panic(err)
	}
	return tss.AwaitReply()
}

func (tss *TestSuiteSession) AwaitReply() *discordgo.Message {
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
	case <-time.After(10 * time.Second):
		panic(errReplyTimeout.Error())
	}
}
