package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/verzac/grocer-discord-bot/config"
	dbUtils "github.com/verzac/grocer-discord-bot/db"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/handlers/api"
	"github.com/verzac/grocer-discord-bot/handlers/slash"
	"github.com/verzac/grocer-discord-bot/monitoring"
	"github.com/verzac/grocer-discord-bot/monitoring/groprometheus"
	"github.com/verzac/grocer-discord-bot/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var db *gorm.DB
var GroBotVersion string = "local"
var BuildTimestamp string = strconv.FormatInt(time.Now().Unix(), 10)
var cw *cloudwatch.CloudWatch
var logger *zap.Logger

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer handlers.Recover(logger)
	if m.Author.ID == s.State.User.ID {
		return
	}
	mLogger := logger.Named("msg.handler")
	mh, err := handlers.NewMessageHandler(s, m, db, GroBotVersion, mLogger)
	if err == handlers.ErrCmdNotProcessable {
		return
	}
	if err != nil {
		// errors shouldn't happen here, but you never know
		mLogger.Error(err.Error())
		return
	}
	metric := monitoring.NewCommandMetric(cw, mh.GetCommand(), mh.GetLogger())
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
	db = dbUtils.Setup(dsn, logger.Named("db"), GroBotVersion)
	services.InitServices(db, logger.Named("service"), d)

	// Set database connection for Prometheus metrics
	groprometheus.SetDB(db)

	// Initialize Prometheus metrics
	groprometheus.InitMetrics()

	// API handler
	go func() {
		if err := api.RegisterAndStart(logger, db); err != nil {
			logger.Error("API returned an error while starting.", zap.Error(err))
		}
	}()
	logger.Info("Setting up discordgo...")
	d.AddHandler(onMessage)

	// Track Discord server count
	d.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger := logger.Named("activity")
		buildTimestampStr, err := strconv.ParseInt(BuildTimestamp, 10, 64)
		if err != nil {
			logger.Error("Cannot parse BuildTimestamp when updating activity status.",
				zap.String("Error", err.Error()),
				zap.String("BuildTimestamp", BuildTimestamp),
			)
			return
		}
		buildTime := time.Unix(buildTimestampStr, 0)
		activityStatusString := fmt.Sprintf(
			"%s (Updated %s)",
			GroBotVersion,
			buildTime.Local().Format("Jan 2"),
		)
		botStatus := "online"
		if config.IsMaintenanceMode() {
			botStatus = "dnd"
			activityStatusString = "Doing maintenance - I'll be back in 10 mins!"
		}
		logger.Info("Updating activity status.",
			zap.String("NewActivity", activityStatusString),
			zap.String("BuildTimestamp", BuildTimestamp),
			zap.String("botStatus", botStatus),
		)
		if err := d.UpdateStatusComplex(discordgo.UpdateStatusData{
			Status: botStatus,
			Activities: []*discordgo.Activity{
				{
					Name: activityStatusString,
					Type: discordgo.ActivityTypeGame,
				},
			},
		}); err != nil {
			logger.Error(err.Error())
		}

		// Update server count metric on ready
		serverCount := handlers.UpdateServerCountMetric(s)
		logger.Info("Updated Discord servers count", zap.Int("serverCount", serverCount))
	})

	// Track when bot joins a server
	d.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		serverCount := handlers.UpdateServerCountMetric(s)
		logger.Info("Bot joined Discord server",
			zap.String("guildID", g.ID),
			zap.String("guildName", g.Name),
			zap.Int("totalServers", serverCount))
	})

	// Track when bot leaves a server
	d.AddHandler(func(s *discordgo.Session, g *discordgo.GuildDelete) {
		serverCount := handlers.UpdateServerCountMetric(s)
		logger.Info("Bot left Discord server",
			zap.String("guildID", g.ID),
			zap.String("guildName", g.Name),
			zap.Int("totalServers", serverCount))
	})

	d.Identify.Intents = discordgo.IntentsGuildMessages
	if err := d.Open(); err != nil {
		panic(err)
	}
	// NOT FATAL SINCE SLASH COMMANDS ARE OPTIONAL
	go func() {
		slashLog := logger.Named("slash")
		defer logPanic()
		slashLog.Info("Starting the slash command registration process...")
		if err := slash.Cleanup(d, slashLog); err != nil {
			slashLog.Error("Cannot cleanup slash commands", zap.Error(err))
			return
		}
		slashLog.Info("Registering slash commands...")
		_, err := slash.Register(d, db, slashLog, GroBotVersion, cw)
		if err != nil {
			slashLog.Error("Cannot register slash commands", zap.Any("Error", err))
			return
		}
		slashLog.Info("Registered slash commands successfully!")
	}()

	logger.Info(
		"Bot is online!",
		zap.String("Version", GroBotVersion),
		zap.Bool("CloudWatchEnabled", monitoring.CloudWatchEnabled()),
		zap.Bool("IsMonitoringEnabled", monitoring.IsMonitoringEnabled()))
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	logger.Info("Shutting down GroceryBot...")
	d.Close()
}
