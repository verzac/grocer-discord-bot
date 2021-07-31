package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	dbUtils "github.com/verzac/grocer-discord-bot/db"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/monitoring"
	"gorm.io/gorm"
)

var db *gorm.DB
var GroBotVersion string = "local"
var cw *cloudwatch.CloudWatch

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	mh, err := handlers.New(s, m, db, GroBotVersion)
	if err == handlers.ErrCmdNotProcessable {
		return
	}
	if err != nil {
		// errors shouldn't happen here, but you never know
		log.Println(fmt.Sprintf("[ERROR] %s", err.Error()))
		return
	}
	metric := monitoring.NewCommandMetric(cw, mh)
	defer metric.Done()
	err = mh.Handle()
	if err != nil {
		log.Println(mh.FmtErrMsg(err))
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load(); err != nil {
		log.Println("Cannot load .env file:", err.Error())
	}
	token := os.Getenv("GROCER_BOT_TOKEN")
	if token == "" {
		panic(errors.New("Cannot get bot token"))
	}
	dsn := os.Getenv("GROCER_BOT_DSN")
	if dsn == "" {
		dsn = "db/gorm.db"
	}
	if monitoring.CloudWatchEnabled() {
		region := "ap-southeast-1"
		sess, err := session.NewSession(&aws.Config{
			Region: &region,
		})
		if err != nil {
			log.Println("[ERROR] Unable to init CW client: " + err.Error())
		} else {
			cw = cloudwatch.New(sess)
		}
	}
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	log.Printf("Using %s\n", dsn)
	db = dbUtils.Setup(dsn)
	log.Println("Setting up discordgo...")
	d.AddHandler(onMessage)
	d.Identify.Intents = discordgo.IntentsGuildMessages
	if err := d.Open(); err != nil {
		panic(err)
	}
	log.Printf("Bot is online! Version=%s CloudWatchEnabled=%t\n", GroBotVersion, monitoring.CloudWatchEnabled())
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	d.Close()
}
