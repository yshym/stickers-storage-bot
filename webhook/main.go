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

// User provides data of users table item
type User struct {
	ID         int
	StickerIDs []string
}

func getUser(ID int) (User, error) {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})

	client := dynamodb.New(sess)

	user := User{}

	result, err := client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				N: aws.String(strconv.Itoa(ID)),
			},
		},
	})
	if err != nil {
		return user, err
	}

	// Return empty slice of stickers if item does not exist
	if result.Item == nil {
		return User{ID: ID, StickerIDs: []string{}}, nil
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func addStickerIDs(ID int, stickerIDs []string) error {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})

	client := dynamodb.New(sess)

	_, err := client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":i": {
				SS: aws.StringSlice(stickerIDs),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				N: aws.String(strconv.Itoa(ID)),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("ADD StickerIDs :i"),
	})
	if err != nil {
		return err
	}

	return err
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

	if update.Message != nil {
		userID := update.Message.From.ID
		stickerID := update.Message.Sticker.FileID

		addStickerIDs(userID, []string{stickerID})
	}

	if update.InlineQuery == nil {
		return okResp, nil
	}

	// Make Query not empty
	update.InlineQuery.Query = "-"

	userID := update.InlineQuery.From.ID

	user, err := getUser(userID)
	if err != nil {
		return badResp, err
	}
	stickerIDs := user.StickerIDs
	resultCachedStickers := []interface{}{}

	for i, stickerID := range stickerIDs {
		resultCachedStickers = append(
			resultCachedStickers,
			tgbotapi.NewInlineQueryResultCachedSticker(
				strconv.Itoa(i),
				stickerID,
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
