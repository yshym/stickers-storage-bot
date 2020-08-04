package db

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Client provides DynamoDB client data
type Client struct {
	DB *dynamodb.DynamoDB
}

// NewClient insantiates new DynamoDB client
func NewClient() (*Client, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})
	if err != nil {
		return &Client{}, err
	}
	database := dynamodb.New(sess)

	return &Client{DB: database}, nil
}

// GetStickers gets a list of stickers by user id
func (client *Client) GetStickers(userID int) ([]Sticker, error) {
	stickers := []Sticker{}

	result, err := client.DB.Query(&dynamodb.QueryInput{
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

// PutSticker inserts a sticker into table
func (client *Client) PutSticker(sticker Sticker) error {
	stickerItem, err := dynamodbattribute.MarshalMap(sticker)
	if err != nil {
		return err
	}

	_, err = client.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Item:      stickerItem,
	})
	if err != nil {
		return err
	}

	return nil
}

// StickerBelongsToUser checks if user has a sticker
func (client *Client) StickerBelongsToUser(userID int, sticker Sticker) (bool, error) {
	result, err := client.DB.Query(&dynamodb.QueryInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		IndexName: aws.String("FileUniqueIDIndex"),
		KeyConditions: map[string]*dynamodb.Condition{
			"FileUniqueID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(sticker.FileUniqueID),
					},
				},
			},
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
		return false, err
	}

	return len(result.Items) != 0, nil
}

// DeleteSticker removes a sticker
func (client *Client) DeleteSticker(sticker Sticker) error {
	_, err := client.DB.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				N: aws.String(strconv.Itoa(sticker.UserID)),
			},
			"FileUniqueID": {
				S: aws.String(sticker.FileUniqueID),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
