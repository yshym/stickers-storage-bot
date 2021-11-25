// Package bot provides stickers storage bot operations and data structures
package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/stickers-storage-bot/db"
)

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
		Results:       resultCachedStickers,
	}

	if _, err := bot.API.AnswerInlineQuery(inlineConf); err != nil {
		return err
	}

	return nil
}

// HandleUpdate handles an update
func (bot *Bot) HandleUpdate(update *tgbotapi.Update) error {
	message := update.Message
	if message != nil && message.Sticker != nil {
		return bot.HandleSticker(message)
	} else if update.InlineQuery != nil {
		return bot.HandleQuery(update.InlineQuery)
	}
	return nil
}

// CheckForUpdates starts checking for updates
func (bot *Bot) CheckForUpdates() {
	updates, err := bot.API.GetUpdatesChan(*bot.UpdateConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	for update := range updates {
		err := bot.HandleUpdate(&update)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}
}