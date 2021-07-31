package handlers

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

const groceryEntryLimit = 100
const maxCmdCharsProcessedBeforeGivingUp = 48

var (
	errCannotConvertInt   = errors.New("Oops, I couldn't see any number there... (ps: you can type !grohelp to get help)")
	errNotValidListNumber = errors.New("Oops, that doesn't seem like a valid list number! (ps: you can type !grohelp to get help)")
	errOverLimit          = errors.New(fmt.Sprintf("Whoops, you've gone over the limit allowed by the bot (max %d grocery entries per server). Please log an issue through GitHub (look at `!grohelp`) to request an increase! Thank you for being a power user! :tada:", groceryEntryLimit))
	errPanic              = errors.New("Hmm... Something broke on my end. Please try again later.")
	errCmdOverLimit       = errors.New(fmt.Sprintf("Command is too long and exceeds the predefined limit (%d).", maxCmdCharsProcessedBeforeGivingUp))
	ErrCmdNotProcessable  = errors.New("Command is not a GroceryBot command.")
)

const CmdPrefix = "!gro"

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

type MessageHandlerContext struct {
	sess           *discordgo.Session
	msg            *discordgo.MessageCreate
	db             *gorm.DB
	grobotVersion  string
	commandContext *CommandContext
}

type CommandContext struct {
	Command        string
	GrocerySublist string
	ArgStr         string
}

func (cc *CommandContext) ToString() string {
	return fmt.Sprintf("<command=%s grocerySublist=%s argStr=%s>", cc.Command, cc.GrocerySublist, cc.ArgStr)
}

func New(sess *discordgo.Session, msg *discordgo.MessageCreate, db *gorm.DB, grobotVersion string) (*MessageHandlerContext, error) {
	cc, err := GetCommandContext(msg.Content)
	if err != nil {
		return nil, err
	}
	return &MessageHandlerContext{
		sess:           sess,
		msg:            msg,
		db:             db,
		grobotVersion:  grobotVersion,
		commandContext: cc,
	}, nil
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
	return fmt.Sprintf("[ERROR] Command=%s GuildID=%s errMsg=%s", m.commandContext.Command, m.msg.GuildID, err.Error())
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
	if mh.commandContext == nil {
		return ""
	}
	return mh.commandContext.Command
}

func GetCommandContext(body string) (*CommandContext, error) {
	if !strings.HasPrefix(body, CmdPrefix) {
		return nil, ErrCmdNotProcessable
	}
	isProcessingSublistLabel := false
	sublistLabel := ""
	command := ""
	loopIndex := 0
	for i, r := range body {
		char := string(r)
		loopIndex = i
		doBreak := false
		switch char {
		case " ", "\n", "\t":
			doBreak = true
			break
		case ":":
			isProcessingSublistLabel = true
		default:
			if isProcessingSublistLabel {
				if len(sublistLabel) >= maxCmdCharsProcessedBeforeGivingUp {
					return nil, errCmdOverLimit
				}
				sublistLabel += char
			} else {
				if len(command) >= maxCmdCharsProcessedBeforeGivingUp {
					return nil, errCmdOverLimit
				}
				command += char
			}
		}
		if doBreak {
			break
		}
	}
	var argStrStartIndex int
	if loopIndex+1 > len(body) {
		argStrStartIndex = len(body)
	} else {
		argStrStartIndex = loopIndex + 1
	}
	commandContext := &CommandContext{
		Command:        command,
		GrocerySublist: sublistLabel,
		ArgStr:         body[argStrStartIndex:],
	}
	return commandContext, nil
}

func (mh *MessageHandlerContext) Recover() {
	if r := recover(); r != nil {
		log.Println(fmt.Sprintf("[PANIC][ERROR] %s\n%s\n", r, debug.Stack()))
		mh.onError(errPanic)
	}
}

func (mh *MessageHandlerContext) Handle() (err error) {
	defer mh.Recover()
	log.Println(mh.commandContext.ToString())
	body := mh.msg.Content
	switch mh.commandContext.Command {
	case CmdGroAdd:
		err = mh.OnAdd()
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
