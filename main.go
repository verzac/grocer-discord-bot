package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/verzac/grocer-discord-bot/handlers"
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var GroBotVersion string

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	body := m.Content
	mh := handlers.New(s, m, db)
	var err error
	if strings.HasPrefix(body, "!gro ") {
		err = mh.OnAdd(strings.TrimPrefix(body, "!gro "))
	} else if strings.HasPrefix(body, "!groremove ") {
		err = mh.OnRemove(strings.TrimPrefix(body, "!groremove "))
	} else if strings.HasPrefix(body, "!groedit ") {
		err = mh.OnEdit(strings.TrimPrefix(body, "!groedit "))
	} else if strings.HasPrefix(body, "!grobulk") {
		err = mh.OnBulk(
			strings.Trim(strings.TrimPrefix(body, "!grobulk"), " \n\t"),
		)
	} else if body == "!grolist" {
		err = mh.OnList()
	} else if body == "!groclear" {
		err = mh.OnClear()
	} else if body == "!grohelp" {
		err = mh.OnHelp(GroBotVersion)
	} else if strings.HasPrefix(body, "!grodeets") {
		err = mh.OnDetail(strings.TrimPrefix(body, "!grodeets "))
	} else if body == "!grohere" {
		err = mh.OnAttach()
	} else if body == "!groreset" {
		err = mh.OnReset()
	}
	if err != nil {
		log.Println(mh.FmtErrMsg(err))
	}
}

func main() {
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
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	log.Printf("Using %s\n", dsn)
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				LogLevel:                  logger.Error, // Log level
				IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,        // Disable color
			},
		),
	})
	if err != nil {
		panic(err)
	}
	log.Println("Auto-Migrating...")
	db.Migrator().AutoMigrate(&models.GroceryEntry{})
	db.Migrator().AutoMigrate(&models.GuildConfig{})
	log.Println("Migration done!")
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
