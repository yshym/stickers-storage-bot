package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yevhenshymotiuk/telegram-lambda-helpers"
)

func main() {
	lambda.Start(helpers.SetWebhook(os.Getenv("TELEGRAM_TOKEN")))
}
