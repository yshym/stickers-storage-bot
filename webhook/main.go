package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/stickers-storage-bot/webhook/stickers"
	"github.com/yevhenshymotiuk/telegram-lambda-helpers/apigateway"
)

func handler(
	request events.APIGatewayProxyRequest,
) (apigateway.Response, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	update := tgbotapi.Update{}

	bodyUnmarshalErr := json.Unmarshal([]byte(request.Body), &update)
	if bodyUnmarshalErr != nil {
		log.Panic(bodyUnmarshalErr)
	}

	message := update.Message
	if message != nil {
		userID := message.From.ID
		sticker := stickers.Sticker{
			UserID:       userID,
			FileUniqueID: message.Sticker.FileUniqueID,
			FileID:       message.Sticker.FileID,
		}

		stickerBelongsToUser, err := sticker.BelongsToUser(userID)
		if err != nil {
			return apigateway.Response404, err
		}

		if stickerBelongsToUser {
			err = sticker.Delete()
		} else {
			err = sticker.Put()
		}
		if err != nil {
			return apigateway.Response404, err
		}
	}

	if update.InlineQuery == nil {
		return apigateway.Response200, nil
	}

	// Make Query not empty
	update.InlineQuery.Query = "-"

	userID := update.InlineQuery.From.ID

	stickers, err := stickers.GetStickers(userID)
	if err != nil {
		return apigateway.Response404, err
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
		log.Println(err)
	}

	return apigateway.Response200, nil
}

func main() {
	lambda.Start(handler)
}
