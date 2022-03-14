// Package bot provides stickers storage bot operations and data structures
package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/yshym/telegram-bot-api/v5"
	"github.com/yshym/stickers-storage-bot/db"
	"github.com/yshym/stickers-storage-bot/helpers"
)

const maxStickers = 50

// Bot provides stickers storage bot data
type Bot struct {
	DBClient     *db.Client
	API          *tgbotapi.BotAPI
	UpdateConfig *tgbotapi.UpdateConfig
}

// NewBot instantiates new stickers storage bot
func NewBot() (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		return nil, err
	}
	log.Printf("Authorized on account %s", api.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	db, err := db.NewClient()
	if err != nil {
		return nil, err
	}

	return &Bot{DBClient: db, API: api, UpdateConfig: &updateConfig}, nil
}

// logUserPrintf logs with id of a user
func logUserPrintf(user *tgbotapi.User, format string, v ...interface{}) {
	idPart := fmt.Sprintf(
		"%s %s (%d): ",
		user.FirstName,
		user.LastName,
		user.ID,
	)
	log.Printf(idPart+format, v...)
}

// HandleSticker handles message with sticker
func (bot *Bot) HandleSticker(message *tgbotapi.Message) error {
	from := message.From
	userID := from.ID
	sticker := db.Sticker{
		UserID:       userID,
		FileUniqueID: message.Sticker.FileUniqueID,
		FileID:       message.Sticker.FileID,
		Timestamp:    helpers.Now().Format("2006-01-02 15:04:05"),
	}

	stickerBelongsToUser, err := bot.DBClient.StickerBelongsToUser(
		userID,
		sticker,
	)
	if err != nil {
		return err
	}

	if stickerBelongsToUser {
		logUserPrintf(from, "Delete sticker '%s'", sticker.FileUniqueID)
		err = bot.DBClient.DeleteSticker(sticker)
	} else {
		stickersCount, err := bot.DBClient.CountStickers(userID)
		if err != nil {
			return err
		}
		if stickersCount == maxStickers {
			msgText := fmt.Sprintf(
				"I am sorry, but only %d stickers can fit into one list",
				maxStickers,
			)
			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ReplyToMessageID = message.MessageID
			bot.API.Send(msg)
			return nil
		}
		logUserPrintf(from, "Put sticker '%s'", sticker.FileUniqueID)
		err = bot.DBClient.PutSticker(sticker)
	}
	if err != nil {
		return err
	}

	return nil
}

// HandleQuery handles inline query
func (bot *Bot) HandleQuery(inlineQuery *tgbotapi.InlineQuery) error {
	// Make Query not empty
	inlineQuery.Query = "-"

	from := inlineQuery.From
	userID := from.ID

	logUserPrintf(from, "Query stickers")
	stickers, err := bot.DBClient.GetStickers(userID)
	if err != nil {
		return err
	}
	resultCachedStickers := []interface{}{}

	for i, sticker := range stickers {
		resultCachedStickers = append(
			resultCachedStickers,
			tgbotapi.NewInlineQueryResultCachedSticker(
				strconv.Itoa(i),
				sticker.FileID,
				"sticker",
			),
		)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       resultCachedStickers,
	}

	if _, err := bot.API.Request(inlineConf); err != nil {
		return err
	}

	return nil
}

// HandleStickerChoice handles sticker choice
func (bot *Bot) HandleStickerChoice(
	choiceInlineResult *tgbotapi.ChosenInlineResult,
) error {
	from := choiceInlineResult.From
	userID := from.ID
	choiceID, err := strconv.Atoi(choiceInlineResult.ResultID)
	if err != nil {
		return err
	}

	stickers, err := bot.DBClient.GetStickers(userID)
	if err != nil {
		return err
	}

	// Choice inline result update got after sticker was deleted
	if choiceID >= len(stickers) {
		return nil
	}
	sticker := stickers[choiceID]
	logUserPrintf(from, "Choose sticker '%s'", sticker.FileUniqueID)

	err = bot.DBClient.UseSticker(sticker)
	if err != nil {
		return err
	}

	return nil
}

// HandleCommand handles 'help' command
func (bot *Bot) HandleHelpCommand(message *tgbotapi.Message) error {
	botUsernameTag := fmt.Sprintf("@%s", bot.API.Self.UserName)
	helpText := fmt.Sprintf("Store stickers:\n"+
		"Send a sticker to save it, send second time to delete\n\n"+
		"View stickers:\n"+
		"- call a bot by typing at sign and its username in the text input field in any chat (%s)\n"+
		"- choose a sticker you want to send\n", botUsernameTag)
	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	msg.Entities = []tgbotapi.MessageEntity{
		{
			Type:   "code",
			Offset: strings.Index(helpText, botUsernameTag),
			Length: len(botUsernameTag),
		},
	}
	_, err := bot.API.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

// HandleCommand handles a command
func (bot *Bot) HandleCommand(message *tgbotapi.Message) error {
	from := message.From
	command := message.Command()

	logUserPrintf(from, "Call command '%s'", command)

	switch command {
	case "help":
		return bot.HandleHelpCommand(message)
	}

	return nil
}

// HandleUpdate handles an update
func (bot *Bot) HandleUpdate(update *tgbotapi.Update) {
	var err error = nil
	message := update.Message
	choiceInlineResult := update.ChosenInlineResult
	if message != nil {
		if message.Sticker != nil {
			err = bot.HandleSticker(message)
		} else if message.IsCommand() {
			err = bot.HandleCommand(message)
		}
	} else if choiceInlineResult != nil {
		err = bot.HandleStickerChoice(choiceInlineResult)
	} else if update.InlineQuery != nil {
		err = bot.HandleQuery(update.InlineQuery)
	}
	if err != nil {
		fmt.Println(err)
	}
}

// CheckForUpdates starts checking for updates
func (bot *Bot) CheckForUpdates() {
	updates := bot.API.GetUpdatesChan(*bot.UpdateConfig)
	for update := range updates {
		go bot.HandleUpdate(&update)
	}
}
