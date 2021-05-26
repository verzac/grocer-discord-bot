package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
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
	body := m.Content
	mh := handlers.New(s, m, db)
	var err error
	cmd := mh.GetCommand()
	metric := monitoring.NewCommandMetric(cw, &mh)
	defer metric.Done()
	switch cmd {
	case handlers.CmdGroAdd:
		err = mh.OnAdd(strings.TrimPrefix(body, "!gro "))
	case handlers.CmdGroRemove:
		err = mh.OnRemove(strings.TrimPrefix(body, "!groremove "))
	case handlers.CmdGroEdit:
		err = mh.OnEdit(strings.TrimPrefix(body, "!groedit "))
	case handlers.CmdGroBulk:
		err = mh.OnBulk(
			strings.Trim(strings.TrimPrefix(body, "!grobulk"), " \n\t"),
		)
	case handlers.CmdGroList:
		err = mh.OnList()
	case handlers.CmdGroClear:
		err = mh.OnClear()
	case handlers.CmdGroHelp:
		err = mh.OnHelp(GroBotVersion)
	case handlers.CmdGroDeets:
		err = mh.OnDetail(strings.TrimPrefix(body, "!grodeets "))
	case handlers.CmdGroHere:
		err = mh.OnAttach()
	case handlers.CmdGroReset:
		err = mh.OnReset()
	}
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
		sess := session.Must(session.NewSession(&aws.Config{
			Region: &region,
		}))
		cw = cloudwatch.New(sess)
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
	log.Println("Bot is online! Version: " + GroBotVersion)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	d.Close()
}
