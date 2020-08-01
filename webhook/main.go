package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/telegram-lambda-helpers/apigateway"
)

// Sticker provides info about sticker
type Sticker struct {
	UserID       int
	FileUniqueID string
	FileID       string
}

func getStickers(userID int) ([]Sticker, error) {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})

	client := dynamodb.New(sess)

	stickers := []Sticker{}

	result, err := client.Query(&dynamodb.QueryInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						N: aws.String(strconv.Itoa(userID)),
					},
				},
			},
		},
	})
	if err != nil {
		return stickers, err
	}

	// Return empty slice of stickers if item does not exist
	if result.Items == nil {
		return stickers, nil
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &stickers)
	if err != nil {
		return stickers, err
	}

	return stickers, nil
}

func (sticker Sticker) put() error {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})

	client := dynamodb.New(sess)

	stickerItem, err := dynamodbattribute.MarshalMap(sticker)
	if err != nil {
		return err
	}

	_, err = client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Item:      stickerItem,
	})
	if err != nil {
		return err
	}

	return nil
}

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
		sticker := Sticker{
			UserID:       message.From.ID,
			FileUniqueID: message.Sticker.FileUniqueID,
			FileID:       message.Sticker.FileID,
		}

		err = sticker.put()
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

	stickers, err := getStickers(userID)
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
