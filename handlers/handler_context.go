package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

const groceryEntryLimit = 100
const maxCmdCharsProcessedBeforeGivingUp = 48

var (
	errCannotConvertInt           = errors.New("Oops, I couldn't see any number there... (ps: you can type !grohelp to get help)")
	errNotValidListNumber         = errors.New("Oops, that doesn't seem like a valid list number! (ps: you can type !grohelp to get help)")
	errOverLimit                  = errors.New(fmt.Sprintf("Whoops, you've gone over the limit allowed by the bot (max %d grocery entries per server). Please log an issue through GitHub (look at `!grohelp`) to request an increase! Thank you for being a power user! :tada:", groceryEntryLimit))
	errPanic                      = errors.New("Hmm... Something broke on my end. Please try again later.")
	errCmdOverLimit               = errors.New(fmt.Sprintf("Command is too long and exceeds the predefined limit (%d).", maxCmdCharsProcessedBeforeGivingUp))
	errGroceryListNotFound        = errors.New("Cannot find grocery list from context.")
	ErrCmdNotProcessable          = errors.New("Command is not a GroceryBot command.")
	ErrMessageSourceNotRecognised = errors.New("No valid message source is detected. ")
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

// Defines the enums to determine where the command is invoked from
const (
	CommandSourceMessageContent = iota
	CommandSourceSlashCommand
)

type MessageHandlerContext struct {
	sess *discordgo.Session
	// msg              *discordgo.MessageCreate
	db               *gorm.DB
	grobotVersion    string
	commandContext   *CommandContext
	logger           *zap.Logger
	groceryEntryRepo repositories.GroceryEntryRepository
	groceryListRepo  repositories.GroceryListRepository
	replyCounter     int
}

type CommandContext struct {
	Command        string
	GrocerySublist string
	ArgStr         string
	GuildID        string
	AuthorID       string
	ChannelID      string
	// see CommandSource* const above
	CommandSourceType int
	// nil if CommandSourceType != CommandSourceSlashCommand, ACCESS SPARINGLY
	Interaction *discordgo.Interaction
}

func (m *MessageHandlerContext) checkReplyCounter() {
	if m.replyCounter > 0 {
		m.logger.Warn(
			"Handler has already replied (this shouldn't happen).",
			append([]zap.Field{zap.Int("ReplyCounter", m.replyCounter)}, m.getDefaultLogFields()...)...,
		)
	}
}

func (m *MessageHandlerContext) reply(msg string) error {
	m.checkReplyCounter()
	switch m.commandContext.CommandSourceType {
	case CommandSourceMessageContent:
		m.replyCounter += 1
		return m.sendMessage(msg)
	case CommandSourceSlashCommand:
		m.replyCounter += 1
		return m.sess.InteractionRespond(m.commandContext.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
	default:
		return ErrMessageSourceNotRecognised
	}
}

func (m *MessageHandlerContext) replyWithEmbed(embed *discordgo.MessageEmbed) error {
	m.checkReplyCounter()
	switch m.commandContext.CommandSourceType {
	case CommandSourceMessageContent:
		m.replyCounter += 1
		_, err := m.sess.ChannelMessageSendEmbed(m.commandContext.ChannelID, embed)
		return err
	case CommandSourceSlashCommand:
		m.replyCounter += 1
		return m.sess.InteractionRespond(m.commandContext.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	default:
		return ErrMessageSourceNotRecognised
	}
}

func (m *MessageHandlerContext) GetGroceryListFromContext() (*models.GroceryList, error) {
	groceryListLabel := m.commandContext.GrocerySublist
	if groceryListLabel != "" {
		groceryList := models.GroceryList{}
		if r := m.db.Where(&models.GroceryList{ListLabel: groceryListLabel, GuildID: m.commandContext.GuildID}).Take(&groceryList); r.Error != nil {
			if r.Error == gorm.ErrRecordNotFound {
				return nil, errGroceryListNotFound
			}
			return nil, r.Error
		}
		return &groceryList, nil
	}
	return nil, nil
}

func (cc *CommandContext) FmtErrInvalidGroceryList() string {
	return fmt.Sprintf("Whoops, I can't seem to find the grocery list labeled as *%s*.", cc.GrocerySublist)
}

func (m *MessageHandlerContext) GetLogger() *zap.Logger {
	return m.logger
}

func (m *MessageHandlerContext) getDefaultLogFields() []zapcore.Field {
	return []zapcore.Field{
		zap.String("Command", m.commandContext.Command),
		zap.String("GuildID", m.commandContext.GuildID),
	}
}

func NewMessageHandler(sess *discordgo.Session, msg *discordgo.MessageCreate, db *gorm.DB, grobotVersion string, logger *zap.Logger) (*MessageHandlerContext, error) {
	cc, err := GetCommandContext(msg.Content, msg.GuildID, msg.Author.ID, msg.ChannelID)
	if err != nil {
		return nil, err
	}
	return &MessageHandlerContext{
		sess: sess,
		// msg:              msg,
		db:               db,
		grobotVersion:    grobotVersion,
		commandContext:   cc,
		logger:           logger,
		groceryEntryRepo: &repositories.GroceryEntryRepositoryImpl{DB: db},
		groceryListRepo:  &repositories.GroceryListRepositoryImpl{DB: db},
	}, nil
}

func NewHandler(sess *discordgo.Session, cc *CommandContext, db *gorm.DB, grobotVersion string, logger *zap.Logger) *MessageHandlerContext {
	return &MessageHandlerContext{
		sess:             sess,
		db:               db,
		grobotVersion:    grobotVersion,
		commandContext:   cc,
		logger:           logger,
		groceryEntryRepo: &repositories.GroceryEntryRepositoryImpl{DB: db},
		groceryListRepo:  &repositories.GroceryListRepositoryImpl{DB: db},
	}
}

// onError handles errors coming in from the handlers and sends the appropriate err resp to the user. returns an error if an error occurs during error-handling; nil otherwise
func (m *MessageHandlerContext) onError(err error) error {
	if discordError, ok := err.(*discordgo.RESTError); ok {
		if discordError.Response.StatusCode == 400 {
			discordErrorResponse := dto.DiscordError{}
			if unmarshalErr := json.Unmarshal(discordError.ResponseBody, &discordErrorResponse); unmarshalErr != nil {
				m.LogError(unmarshalErr)
			} else if discordErrorResponse.Code == 50035 {
				maxLengthExceeded := false
				for _, e := range discordErrorResponse.Errors.Content.Errors {
					if e.Code == "BASE_TYPE_MAX_LENGTH" {
						maxLengthExceeded = true
					}
				}
				if maxLengthExceeded {
					m.logger.Info("Max length for message sending exceeded.", m.getDefaultLogFields()...)
					// not a big deal, tell the user off
					if sErr := m.reply(":exploding_head: Whoops, we can't send you a reply because the reply is going to be too big! Do try clearing your grocery lists or make your items shorter, as I can only send messages (e.g. grocery lists) which are below 2000 chars."); sErr != nil {
						m.LogError(errors.Wrap(sErr, "Cannot send message to notify the caller that the message is too long."))
					}
					return nil
				}
			}
		}
	}
	m.LogError(err)
	_, sErr := m.sess.ChannelMessageSend(m.commandContext.ChannelID, fmt.Sprintf(":helmet_with_cross: Oops, something broke! Give it a day or so and it'll be fixed by the team (or you can follow up this issue with us at our Discord server!). Error:\n```\n%s\n```", err.Error()))
	if sErr != nil {
		wrappedErr := errors.Wrap(err, sErr.Error())
		m.LogError(wrappedErr)
		return wrappedErr
	}
	return nil // mark it as handled
}

func fmtItemNotFoundErrorMsg(itemIndex int) string {
	return fmt.Sprintf("Hmm... Can't seem to find item #%d on the list :/", itemIndex)
}

func (m *MessageHandlerContext) LogError(err error, extraFields ...zapcore.Field) {
	m.GetLogger().Error(
		err.Error(),
		append(
			m.getDefaultLogFields(),
			extraFields...,
		)...,
	)
}

func (m *MessageHandlerContext) onItemNotFound(itemIndex int) error {
	err := m.sendMessage(fmtItemNotFoundErrorMsg(itemIndex))
	if err != nil {
		return m.onError(err)
	}
	return m.displayList()
}

func (m *MessageHandlerContext) sendMessage(msg string) error {
	_, sErr := m.sess.ChannelMessageSendComplex(m.commandContext.ChannelID, &discordgo.MessageSend{
		Content: msg,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			// do not allow mentions by default
			Parse: []discordgo.AllowedMentionType{},
		},
	})
	return sErr
}

func (m *MessageHandlerContext) onGetGroceryListError(err error) error {
	switch err {
	case errGroceryListNotFound:
		return m.sendMessage(m.commandContext.FmtErrInvalidGroceryList())
	default:
		return m.onError(err)
	}
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

func GetCommandContext(body string, guildID string, authorID string, channelID string) (*CommandContext, error) {
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
			if isProcessingSublistLabel == true {
				return nil, ErrCmdNotProcessable
			}
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
		Command:           command,
		GrocerySublist:    sublistLabel,
		ArgStr:            strings.TrimLeft(body[argStrStartIndex:], "\n "),
		GuildID:           guildID,
		AuthorID:          authorID,
		ChannelID:         channelID,
		CommandSourceType: CommandSourceMessageContent,
	}
	return commandContext, nil
}

func (m *MessageHandlerContext) Recover() {
	if r := recover(); r != nil {
		m.GetLogger().Error(
			"Panic encountered! Recovering...",
			zap.Any("Panic", r),
			zap.String("Command", m.commandContext.Command),
			zap.String("GuildID", m.commandContext.GuildID),
			zap.Stack("Stack"),
		)
		m.onError(errPanic)
	}
}

func Recover(logger *zap.Logger) {
	if r := recover(); r != nil {
		logger.Error(
			"Very, very bad panic encountered! Recovering...",
			zap.Any("Panic", r),
			zap.Stack("Stack"),
		)
	}
}

func (mh *MessageHandlerContext) Handle() (err error) {
	defer mh.Recover()
	mh.GetLogger().Debug(
		"Handling command.",
		zap.String("Command", mh.commandContext.Command),
		zap.String("ArgStr", mh.commandContext.ArgStr),
		zap.String("GrocerySublist", mh.commandContext.GrocerySublist),
	)
	switch mh.commandContext.Command {
	case CmdGroAdd:
		err = mh.OnAdd()
	case CmdGroRemove:
		err = mh.OnRemove()
	case CmdGroEdit:
		err = mh.OnEdit()
	case CmdGroBulk:
		err = mh.OnBulk()
	case CmdGroList:
		err = mh.OnList()
	case CmdGroClear:
		err = mh.OnClear()
	case CmdGroHelp:
		err = mh.OnHelp()
	case CmdGroDeets:
		err = mh.OnDetail()
	case CmdGroHere:
		err = mh.OnAttach()
	case CmdGroReset:
		err = mh.OnReset()
	default:
		err = ErrCmdNotProcessable
	}
	if err != nil && err != ErrCmdNotProcessable {
		return mh.onError(err)
	}
	return err
}
