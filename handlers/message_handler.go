package handlers

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

var (
	errCannotConvertInt   = errors.New("Oops, I couldn't see any number there... (ps: you can type !grohelp to get help)")
	errNotValidListNumber = errors.New("Oops, that doesn't seem like a valid list number! (ps: you can type !grohelp to get help)")
	errOverLimit          = errors.New(fmt.Sprintf("Whoops, you've gone over the limit allowed by the bot (max %d grocery entries per server). Please log an issue through GitHub (look at `!grohelp`) to request an increase! Thank you for being a power user! :tada:", groceryEntryLimit))
)

const groceryEntryLimit = 100

type MessageHandler struct {
	sess *discordgo.Session
	msg  *discordgo.MessageCreate
	db   *gorm.DB
}

func New(sess *discordgo.Session, msg *discordgo.MessageCreate, db *gorm.DB) MessageHandler {
	return MessageHandler{sess: sess, msg: msg, db: db}
}

func (m *MessageHandler) onError(err error) error {
	log.Println(m.fmtErrMsg(err))
	_, sErr := m.sess.ChannelMessageSend(m.msg.ChannelID, fmt.Sprintf("Oops! Something broke:\n%s", err.Error()))
	if sErr != nil {
		log.Println("Unable to send error message.", m.fmtErrMsg(err))
	}
	return err
}

func fmtItemNotFoundErrorMsg(itemIndex int) string {
	return fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex)
}

func (m *MessageHandler) checkLimit(guildID string, newItemCount int64) error {
	var count int64
	if r := m.db.Model(&models.GroceryEntry{}).Where(&models.GroceryEntry{GuildID: guildID}).Count(&count); r.Error != nil {
		return r.Error
	}
	if count+newItemCount > groceryEntryLimit {
		return errOverLimit
	}
	return nil
}

func (m *MessageHandler) fmtErrMsg(err error) string {
	return fmt.Sprintf("GuildID=%s errMsg=%s", m.msg.GuildID, err.Error())
}

func (m *MessageHandler) sendMessage(msg string) error {
	_, sErr := m.sess.ChannelMessageSendComplex(m.msg.ChannelID, &discordgo.MessageSend{
		Content: msg,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			// do not allow mentions by default
			Parse: []discordgo.AllowedMentionType{},
		},
	})
	if sErr != nil {
		log.Println("Unable to send message.", m.fmtErrMsg(sErr))
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

func prettyItemIndexList(itemIndexes []int) string {
	tokens := make([]string, len(itemIndexes))
	for i, itemIndex := range itemIndexes {
		tokens[i] = fmt.Sprintf("#%d", itemIndex)
	}
	return strings.Join(tokens, ", ")
}

func prettyItems(gList []models.GroceryEntry) string {
	tokens := make([]string, len(gList))
	for i, gEntry := range gList {
		format := "*%s*"
		if i == len(gList)-1 && len(gList) > 1 {
			format = fmt.Sprintf("and %s", format)
		}
		tokens[i] = fmt.Sprintf(format, gEntry.ItemDesc)
	}
	return strings.Join(tokens, ", ")
}
