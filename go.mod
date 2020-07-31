module github.com/yevhenshymotiuk/stickers-storage-bot

go 1.14

require (
	github.com/aws/aws-lambda-go v1.18.0
	github.com/aws/aws-sdk-go v1.32.3
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/yevhenshymotiuk/telegram-lambda-helpers v0.0.0-20200729200603-7045ad4169de
)

replace github.com/go-telegram-bot-api/telegram-bot-api => github.com/go-telegram-bot-api/telegram-bot-api v1.0.1-0.20200729154208-fb8759e91dfc
