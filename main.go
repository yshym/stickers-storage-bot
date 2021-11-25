package main

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/stickers-storage-bot/db"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	client, err := db.NewClient()
	if err != nil {
		log.Panic(err)
	}

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		message := update.Message
		if message != nil {
			if message.Sticker == nil {
				continue
			}
			userID := message.From.ID
			sticker := db.Sticker{
				UserID:       userID,
				FileUniqueID: message.Sticker.FileUniqueID,
				FileID:       message.Sticker.FileID,
			}

			stickerBelongsToUser, err := client.StickerBelongsToUser(
				userID,
				sticker,
			)
			if err != nil {
				log.Panic(err)
			}

			if stickerBelongsToUser {
				err = client.DeleteSticker(sticker)
			} else {
				err = client.PutSticker(sticker)
			}
			if err != nil {
				log.Panic(err)
			}
		}
		if update.InlineQuery == nil {
			continue
		}

		// Make Query not empty
		update.InlineQuery.Query = "-"

		userID := update.InlineQuery.From.ID

		stickers, err := client.GetStickers(userID)
		if err != nil {
			log.Panic(err)
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
			InlineQueryID: update.InlineQuery.ID,
			Results:       resultCachedStickers,
		}

		if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
			log.Panic(err)
		}
	}
}
