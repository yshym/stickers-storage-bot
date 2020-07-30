package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/telegram-lambda-helpers/apigateway"
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

	if update.InlineQuery == nil {
		return okResp, nil
	}

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

	return okResp, nil
}

func main() {
	lambda.Start(handler)
}
