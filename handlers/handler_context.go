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

// Note: make sure this is alphabetically ordered so that we don't get confused
const (
	CmdGroAdd    = "!gro"
	CmdGroBulk   = "!grobulk"
	CmdGroClear  = "!groclear"
	CmdGroDeets  = "!grodeets"
	CmdGroEdit   = "!groedit"
	CmdGroHelp   = "!grohelp"
	CmdGroHere   = "!grohere"
	CmdGroList   = "!grolist"
	CmdGroRemove = "!groremove"
	CmdGroReset  = "!groreset"
)

const groceryEntryLimit = 100

type MessageHandlerContext struct {
	sess          *discordgo.Session
	msg           *discordgo.MessageCreate
	db            *gorm.DB
	grobotVersion string
}

func New(sess *discordgo.Session, msg *discordgo.MessageCreate, db *gorm.DB, grobotVersion string) MessageHandlerContext {
	return MessageHandlerContext{sess: sess, msg: msg, db: db, grobotVersion: grobotVersion}
}

func (m *MessageHandlerContext) onError(err error) error {
	log.Println(m.FmtErrMsg(err))
	_, sErr := m.sess.ChannelMessageSend(m.msg.ChannelID, fmt.Sprintf("Oops! Something broke:\n%s", err.Error()))
	if sErr != nil {
		log.Println("Unable to send error message.", m.FmtErrMsg(err))
	}
	return err
}

func fmtItemNotFoundErrorMsg(itemIndex int) string {
	return fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex)
}

func (m *MessageHandlerContext) checkLimit(guildID string, newItemCount int64) error {
	var count int64
	if r := m.db.Model(&models.GroceryEntry{}).Where(&models.GroceryEntry{GuildID: guildID}).Count(&count); r.Error != nil {
		return r.Error
	}
	if count+newItemCount > groceryEntryLimit {
		return errOverLimit
	}
	return nil
}

func (m *MessageHandlerContext) FmtErrMsg(err error) string {
	return fmt.Sprintf("[ERROR] Command=%s GuildID=%s errMsg=%s", m.GetCommand(), m.msg.GuildID, err.Error())
}

func (m *MessageHandlerContext) sendMessage(msg string) error {
	_, sErr := m.sess.ChannelMessageSendComplex(m.msg.ChannelID, &discordgo.MessageSend{
		Content: msg,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			// do not allow mentions by default
			Parse: []discordgo.AllowedMentionType{},
		},
	})
	if sErr != nil {
		log.Println("Unable to send message.", m.FmtErrMsg(sErr))
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

func (mh *MessageHandlerContext) GetCommand() string {
	body := mh.msg.Content
	if strings.HasPrefix(body, "!gro ") {
		return CmdGroAdd
	} else if strings.HasPrefix(body, "!groremove ") {
		return CmdGroRemove
	} else if strings.HasPrefix(body, "!groedit ") {
		return CmdGroEdit
	} else if strings.HasPrefix(body, "!grobulk") {
		return CmdGroBulk
	} else if body == "!grolist" {
		return CmdGroList
	} else if body == "!groclear" {
		return CmdGroClear
	} else if body == "!grohelp" {
		return CmdGroHelp
	} else if strings.HasPrefix(body, "!grodeets") {
		return CmdGroDeets
	} else if body == "!grohere" {
		return CmdGroHere
	} else if body == "!groreset" {
		return CmdGroReset
	} else {
		return ""
	}
}

func (mh *MessageHandlerContext) Handle() (err error) {
	body := mh.msg.Content
	switch mh.GetCommand() {
	case CmdGroAdd:
		err = mh.OnAdd(strings.TrimPrefix(body, "!gro "))
	case CmdGroRemove:
		err = mh.OnRemove(strings.TrimPrefix(body, "!groremove "))
	case CmdGroEdit:
		err = mh.OnEdit(strings.TrimPrefix(body, "!groedit "))
	case CmdGroBulk:
		err = mh.OnBulk(
			strings.Trim(strings.TrimPrefix(body, "!grobulk"), " \n\t"),
		)
	case CmdGroList:
		err = mh.OnList()
	case CmdGroClear:
		err = mh.OnClear()
	case CmdGroHelp:
		err = mh.OnHelp(mh.grobotVersion)
	case CmdGroDeets:
		err = mh.OnDetail(strings.TrimPrefix(body, "!grodeets "))
	case CmdGroHere:
		err = mh.OnAttach()
	case CmdGroReset:
		err = mh.OnReset()
	}
	return err
}
