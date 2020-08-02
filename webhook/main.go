package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/telegram-lambda-helpers/apigateway"
	"github.com/yevhenshymotiuk/stickers-storage-bot/webhook/stickers"
)

var (
	okResp = apigateway.Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            "Ok",
	}
	badResp = apigateway.Response{StatusCode: 404}
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
		sticker := stickers.Sticker{
			UserID:       message.From.ID,
			FileUniqueID: message.Sticker.FileUniqueID,
			FileID:       message.Sticker.FileID,
		}

		err = sticker.Put()
		if err != nil {
			return badResp, err
		}
	}

	if update.InlineQuery == nil {
		return okResp, nil
	}

	// Make Query not empty
	update.InlineQuery.Query = "-"

	userID := update.InlineQuery.From.ID

	stickers, err := stickers.GetStickers(userID)
	if err != nil {
		return badResp, err
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

	return okResp, nil
}

func main() {
	lambda.Start(handler)
}
