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

func fmtItemNotFoundErrorMsg(itemIndex int) string {
	return fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex)
}

var (
	errCannotConvertInt   = errors.New("Oops, I couldn't see any number there... (ps: you can type !grohelp to get help)")
	errNotValidListNumber = errors.New("Oops, that doesn't seem like a valid list number! (ps: you can type !grohelp to get help)")
)

var db *gorm.DB

type groceryEntry struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	ItemDesc  string `gorm:"not null"`
	GuildID   string `gorm:"index;not null"`
	CreatorID *string
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

func sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	_, sErr := s.ChannelMessageSend(m.ChannelID, msg)
	if sErr != nil {
		log.Println("Unable to send message.", sErr.Error())
	}
	return sErr
}

func onError(s *discordgo.Session, m *discordgo.MessageCreate, err error) error {
	_, sErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Oops! Something broke:\n%s", err.Error()))
	if sErr != nil {
		log.Println("Unable to send error message.", err.Error())
	}
	return err
}

func onAdd(argStr string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	if r := db.Create(&groceryEntry{ItemDesc: argStr, GuildID: m.GuildID}); r.Error != nil {
		return onError(s, m, r.Error)
	}
	return sendMessage(s, m, fmt.Sprintf("Added *%s* into your grocery list!", argStr))
}

func onList(s *discordgo.Session, m *discordgo.MessageCreate) error {
	groceries := make([]groceryEntry, 0)
	if r := db.Where(&groceryEntry{GuildID: m.GuildID}).Find(&groceries); r.Error != nil {
		return onError(s, m, r.Error)
	}
	msg := "Here's your grocery list:\n"
	for i, grocery := range groceries {
		msg += fmt.Sprintf("%d: %s\n", i+1, grocery.ItemDesc)
	}
	return sendMessage(s, m, msg)
}

func onClear(s *discordgo.Session, m *discordgo.MessageCreate) error {
	r := db.Delete(groceryEntry{}, "guild_id = ?", m.GuildID)
	if r.Error != nil {
		return onError(s, m, r.Error)
	}
	msg := fmt.Sprintf("Deleted %d items off your grocery list!", r.RowsAffected)
	return sendMessage(s, m, msg)
}

func onRemove(argStr string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	itemIndex, err := toItemIndex(argStr)
	if err != nil {
		return sendMessage(s, m, err.Error())
	}
	g := groceryEntry{}
	r := db.Where("guild_id = ?", m.GuildID).Offset(itemIndex - 1).First(&g)
	if r.Error != nil {
		if r.Error == gorm.ErrRecordNotFound {
			sendMessage(s, m, fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex))
			return onList(s, m)
		}
		return onError(s, m, r.Error)
	}
	if r.RowsAffected == 0 {
		msg := fmt.Sprintf("Cannot find item with index %d!", itemIndex)
		return sendMessage(s, m, msg)
	}
	db.Delete(g)
	return sendMessage(s, m, fmt.Sprintf("Deleted *%s* off your grocery list!", g.ItemDesc))
}

func onEdit(argStr string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	argTokens := strings.SplitN(argStr, " ", 2)
	if len(argTokens) != 2 {
		return sendMessage(s, m, fmt.Sprintf("Oops, I can't seem to understand you. Perhaps try typing **!groedit 1 Whatever you want the name of this entry to be**?"))
	}
	itemIndex, err := toItemIndex(argTokens[0])
	if err != nil {
		return sendMessage(s, m, err.Error())
	}
	newItemDesc := argTokens[1]
	g := groceryEntry{}
	fr := db.Where("guild_id = ?", m.GuildID).Offset(itemIndex - 1).First(&g)
	if fr.Error != nil {
		if errors.Is(fr.Error, gorm.ErrRecordNotFound) {
			sendMessage(s, m, fmtItemNotFoundErrorMsg(itemIndex))
			return onList(s, m)
		}
		return onError(s, m, fr.Error)
	}
	if fr.RowsAffected == 0 {
		msg := fmt.Sprintf("Cannot find item with index %d!", itemIndex)
		return sendMessage(s, m, msg)
	}
	g.ItemDesc = newItemDesc
	if sr := db.Save(g); sr.Error != nil {
		log.Println(sr.Error)
		return sendMessage(s, m, "Welp, something went wrong while saving. Please try again :)")
	}
	return sendMessage(s, m, fmt.Sprintf("Updated item #%d on your grocery list to *%s*", itemIndex, g.ItemDesc))
}

func onBulk(argStr string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	items := strings.Split(
		strings.Trim(argStr, "\n \t"),
		"\n",
	)
	toInsert := make([]groceryEntry, len(items))
	for i, item := range items {
		aID := m.Author.ID
		toInsert[i] = groceryEntry{
			ItemDesc:  strings.Trim(item, " \n\t"),
			GuildID:   m.GuildID,
			CreatorID: &aID,
		}
	}
	r := db.Create(&toInsert)
	if r.Error != nil {
		log.Println(r.Error)
		return sendMessage(s, m, "Hmm... Cannot save your grocery list. Please try again later :)")
	}
	return sendMessage(s, m, fmt.Sprintf("Added %d items into your list!", r.RowsAffected))
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	body := m.Content
	if strings.HasPrefix(body, "!gro ") {
		onAdd(strings.TrimPrefix(body, "!gro "), s, m)
	} else if strings.HasPrefix(body, "!groremove ") {
		onRemove(strings.TrimPrefix(body, "!groremove "), s, m)
	} else if strings.HasPrefix(body, "!groedit ") {
		onEdit(strings.TrimPrefix(body, "!groedit "), s, m)
	} else if strings.HasPrefix(body, "!grobulk") {
		onBulk(
			strings.Trim(strings.TrimPrefix(body, "!grobulk"), " \n\t"),
			s,
			m,
		)
	} else if body == "!grolist" {
		onList(s, m)
	} else if body == "!groclear" {
		onClear(s, m)
	} else if body == "!grohelp" {
		onHelp(s, m)
	}
	// qty
	// !gro Orange Juice
	// !groclear
	// !grolist
	// !groremove
}

func onHelp(s *discordgo.Session, m *discordgo.MessageCreate) error {
	return sendMessage(s, m,
		`!grohelp: get help!
!gro <name>: adds an item to your grocery list
!groremove <n>: removes item #n from your grocery list
!grolist: list all the groceries in your grocery list
!groclear: clears your grocery list
!groedit <n> <new name>: updates item #n to a new name/entry

You can also do !grobulk to add your own grocery list. Format:

`+"```"+`
!grobulk
eggs
Soap 1pc
Liquid soap 500 ml
`+"```"+`

These 3 new items will be added to your existing grocery list!
`,
	)
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
