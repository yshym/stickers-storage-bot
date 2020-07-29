package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
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

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		fmt.Println(update)
		if update.InlineQuery != nil {
			update.InlineQuery.Query = "-"

			sticker := tgbotapi.NewInlineQueryResultCachedSticker(
				"sticker1",
				"CAACAgQAAxkBAAMIXyGRJmVyxqSoowUUBy1HqT_SGKAAAgUAA6MaLxnmpO8KYecXHxoE",
				"doge",
			)

			inlineConf := tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				Results:       []interface{}{sticker},
			}

			if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
				log.Println(err)
			}
		}
	}
}
