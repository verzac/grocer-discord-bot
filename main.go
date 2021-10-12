package main

import (
	"errors"
	"fmt"
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
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var db *gorm.DB
var GroBotVersion string = "local"
var cw *cloudwatch.CloudWatch
var logger *zap.Logger

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer handlers.Recover(logger)
	if m.Author.ID == s.State.User.ID {
		return
	}
	mh, err := handlers.New(s, m, db, GroBotVersion, logger)
	if err == handlers.ErrCmdNotProcessable {
		return
	}
	if err != nil {
		// errors shouldn't happen here, but you never know
		logger.Error(err.Error())
		return
	}
	metric := monitoring.NewCommandMetric(cw, mh)
	err = mh.Handle()
	if err == handlers.ErrCmdNotProcessable {
		return
	}
	metric.Done()
	if err != nil {
		mh.LogError(err)
	}
}

func logPanic() {
	if r := recover(); r != nil && logger != nil {
		logger.Error(
			"Fatal panic",
			zap.Any("Panic", r),
			zap.String("Version", GroBotVersion),
		)
		os.Exit(1)
	}
}

func main() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	if GroBotVersion == "local" {
		l, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		logger = l
	} else {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		logger = l
	}
	defer logPanic()
	defer logger.Sync()
	if err := godotenv.Load(); err != nil {
		logger.Debug("Cannot load .env file: " + err.Error())
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
			logger.Error("Unable to init CW client: " + err.Error())
		} else {
			cw = cloudwatch.New(sess)
		}
	}
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	logger.Info(fmt.Sprintf("Using %s\n", dsn))
	db = dbUtils.Setup(dsn, logger, GroBotVersion)
	logger.Info("Setting up discordgo...")
	d.AddHandler(onMessage)
	d.Identify.Intents = discordgo.IntentsGuildMessages
	if err := d.Open(); err != nil {
		panic(err)
	}
	logger.Info(
		"Bot is online!",
		zap.String("Version", GroBotVersion),
		zap.Bool("CloudWatchEnabled", monitoring.CloudWatchEnabled()),
		zap.Bool("IsMonitoringEnabled", monitoring.IsMonitoringEnabled()))
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	logger.Info("Shutting down GroceryBot...")
	d.Close()
}
