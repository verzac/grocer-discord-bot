package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const cmdPrefix = "!gro "

var (
	errCannotConvertInt   = errors.New("Oops, I couldn't see any number there... (ps: you can type !grohelp to get help)")
	errNotValidListNumber = errors.New("Oops, that doesn't seem like a valid list number! (ps: you can type !grohelp to get help)")
)

var db *gorm.DB

type groceryEntry struct {
	ID          uint `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ItemDesc    string `gorm:"not null"`
	GuildID     string `gorm:"index;not null"`
	UpdatedByID *string
}

func fmtItemNotFoundErrorMsg(itemIndex int) string {
	return fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex)
}

type messageHandler struct {
	sess *discordgo.Session
	msg  *discordgo.MessageCreate
}

func (m *messageHandler) sendMessage(msg string) error {
	_, sErr := m.sess.ChannelMessageSendComplex(m.msg.ChannelID, &discordgo.MessageSend{
		Content: msg,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			// do not allow mentions by default
			Parse: []discordgo.AllowedMentionType{},
		},
	})
	if sErr != nil {
		log.Println("Unable to send message.", sErr.Error())
	}
	return sErr
}

func toItemIndex(argStr string) (int, error) {
	itemIndex, err := strconv.Atoi(argStr)
	if err != nil {
		return 0, errCannotConvertInt
	}
	if itemIndex < 1 {
		return 0, errNotValidListNumber
	}
	return itemIndex, nil
}

func (m *messageHandler) onError(err error) error {
	_, sErr := m.sess.ChannelMessageSend(m.msg.ChannelID, fmt.Sprintf("Oops! Something broke:\n%s", err.Error()))
	if sErr != nil {
		log.Println("Unable to send error message.", err.Error())
	}
	return err
}

func (m *messageHandler) OnAdd(argStr string) error {
	if r := db.Create(&groceryEntry{ItemDesc: argStr, GuildID: m.msg.GuildID, UpdatedByID: &m.msg.Author.ID}); r.Error != nil {
		return m.onError(r.Error)
	}
	return m.sendMessage(fmt.Sprintf("Added *%s* into your grocery list!", argStr))
}

func (m *messageHandler) OnList() error {
	groceries := make([]groceryEntry, 0)
	if r := db.Where(&groceryEntry{GuildID: m.msg.GuildID}).Find(&groceries); r.Error != nil {
		return m.onError(r.Error)
	}
	msg := "Here's your grocery list:\n"
	for i, grocery := range groceries {
		// updatedByString := ""
		// if grocery.UpdatedByID != nil {
		// 	updatedByString = fmt.Sprintf(" (<@%s>)", *grocery.UpdatedByID)
		// }
		msg += fmt.Sprintf("%d: %s\n", i+1, grocery.ItemDesc)
	}
	return m.sendMessage(msg)
}

func (m *messageHandler) OnClear() error {
	r := db.Delete(groceryEntry{}, "guild_id = ?", m.msg.GuildID)
	if r.Error != nil {
		return m.onError(r.Error)
	}
	msg := fmt.Sprintf("Deleted %d items off your grocery list!", r.RowsAffected)
	return m.sendMessage(msg)
}

func (m *messageHandler) OnRemove(argStr string) error {
	itemIndex, err := toItemIndex(argStr)
	if err != nil {
		return m.sendMessage(err.Error())
	}
	g := groceryEntry{}
	r := db.Where("guild_id = ?", m.msg.GuildID).Offset(itemIndex - 1).First(&g)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			m.sendMessage(fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex))
			return m.OnList()
		}
		return m.onError(r.Error)
	}
	if r.RowsAffected == 0 {
		msg := fmt.Sprintf("Cannot find item with index %d!", itemIndex)
		return m.sendMessage(msg)
	}
	db.Delete(g)
	return m.sendMessage(fmt.Sprintf("Deleted *%s* off your grocery list!", g.ItemDesc))
}

func (m *messageHandler) OnEdit(argStr string) error {
	argTokens := strings.SplitN(argStr, " ", 2)
	if len(argTokens) != 2 {
		return m.sendMessage(fmt.Sprintf("Oops, I can't seem to understand you. Perhaps try typing **!groedit 1 Whatever you want the name of this entry to be**?"))
	}
	itemIndex, err := toItemIndex(argTokens[0])
	if err != nil {
		return m.sendMessage(err.Error())
	}
	newItemDesc := argTokens[1]
	g := groceryEntry{}
	fr := db.Where("guild_id = ?", m.msg.GuildID).Offset(itemIndex - 1).First(&g)
	if fr.Error != nil {
		if errors.Is(fr.Error, gorm.ErrRecordNotFound) {
			m.sendMessage(fmtItemNotFoundErrorMsg(itemIndex))
			return m.OnList()
		}
		return m.onError(fr.Error)
	}
	if fr.RowsAffected == 0 {
		msg := fmt.Sprintf("Cannot find item with index %d!", itemIndex)
		return m.sendMessage(msg)
	}
	g.ItemDesc = newItemDesc
	g.UpdatedByID = &m.msg.Author.ID
	if sr := db.Save(g); sr.Error != nil {
		log.Println(sr.Error)
		return m.sendMessage("Welp, something went wrong while saving. Please try again :)")
	}
	return m.sendMessage(fmt.Sprintf("Updated item #%d on your grocery list to *%s*", itemIndex, g.ItemDesc))
}

func (m *messageHandler) OnBulk(argStr string) error {
	items := strings.Split(
		strings.Trim(argStr, "\n \t"),
		"\n",
	)
	toInsert := make([]groceryEntry, len(items))
	for i, item := range items {
		aID := m.msg.Author.ID
		toInsert[i] = groceryEntry{
			ItemDesc:    strings.Trim(item, " \n\t"),
			GuildID:     m.msg.GuildID,
			UpdatedByID: &aID,
		}
	}
	r := db.Create(&toInsert)
	if r.Error != nil {
		log.Println(r.Error)
		return m.sendMessage("Hmm... Cannot save your grocery list. Please try again later :)")
	}
	return m.sendMessage(fmt.Sprintf("Added %d items into your list!", r.RowsAffected))
}

func (m *messageHandler) OnHelp() error {
	return m.sendMessage(
		`!grohelp: get help!
!gro <name>: adds an item to your grocery list
!groremove <n>: removes item #n from your grocery list
!grolist: list all the groceries in your grocery list
!groclear: clears your grocery list
!groedit <n> <new name>: updates item #n to a new name/entry

You can also do !grobulk to add your own grocery list. Format:

` + "```" + `
!grobulk
eggs
Soap 1pc
Liquid soap 500 ml
` + "```" + `

These 3 new items will be added to your existing grocery list!

For more information and/or any other questions, you can get in touch with my maintainer through my GitHub repo: https://github.com/verzac/grocer-discord-bot
`,
	)
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	body := m.Content
	mh := messageHandler{sess: s, msg: m}
	if strings.HasPrefix(body, "!gro ") {
		mh.OnAdd(strings.TrimPrefix(body, "!gro "))
	} else if strings.HasPrefix(body, "!groremove ") {
		mh.OnRemove(strings.TrimPrefix(body, "!groremove "))
	} else if strings.HasPrefix(body, "!groedit ") {
		mh.OnEdit(strings.TrimPrefix(body, "!groedit "))
	} else if strings.HasPrefix(body, "!grobulk") {
		mh.OnBulk(
			strings.Trim(strings.TrimPrefix(body, "!grobulk"), " \n\t"),
		)
	} else if body == "!grolist" {
		mh.OnList()
	} else if body == "!groclear" {
		mh.OnClear()
	} else if body == "!grohelp" {
		mh.OnHelp()
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
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Println("Auto-Migrating...")
	db.Migrator().AutoMigrate(&groceryEntry{})
	log.Println("Migration done!")
	log.Println("Setting up discordgo...")
	d.AddHandler(onMessage)
	d.Identify.Intents = discordgo.IntentsGuildMessages
	if err := d.Open(); err != nil {
		panic(err)
	}
	log.Println("Bot is online!")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	d.Close()
}
