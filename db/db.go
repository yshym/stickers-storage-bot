package db

import (
	"os"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
	creds := credentials.NewEnvCredentials()
	creds.Get()
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-north-1"),
		Endpoint:    aws.String("http://dynamodb:8000"),
		Credentials: creds,
	})
	if err != nil {
		return &Client{}, err
	}
	database := dynamodb.New(sess)

	return &Client{DB: database}, nil
}

// CountStickers counts user's stickers
func (client *Client) CountStickers(userID int64) (int64, error) {
	result, err := client.DB.Scan(&dynamodb.ScanInput{
		TableName:        aws.String(os.Getenv("DYNAMODB_TABLE")),
		FilterExpression: aws.String("#f = :v"),
		ExpressionAttributeNames: map[string]*string{
			"#f": aws.String("UserID"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				N: aws.String(strconv.Itoa(int(userID))),
			},
		},
	})
	return *result.Count, err
}

// GetStickers gets a list of stickers by user id
func (client *Client) GetStickers(userID int64) (Stickers, error) {
	stickers := Stickers{}

	result, err := client.DB.Query(&dynamodb.QueryInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						N: aws.String(strconv.Itoa(int(userID))),
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

	sort.Sort(stickers)

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

// UseSticker increments UseCount of a sticker
func (client *Client) UseSticker(sticker Sticker) error {
	_, err := client.DB.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				N: aws.String(strconv.Itoa(int(sticker.UserID))),
			},
			"FileUniqueID": {
				S: aws.String(sticker.FileUniqueID),
			},
		},
		UpdateExpression: aws.String("set UseCount = UseCount + :num"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":num": {
				N: aws.String("1"),
			},
		},
		ReturnValues: aws.String("NONE"),
	})
	if err != nil {
		return err
	}

	return nil
}

// StickerBelongsToUser checks if user has a sticker
func (client *Client) StickerBelongsToUser(
	userID int64,
	sticker Sticker,
) (bool, error) {
	result, err := client.DB.Query(&dynamodb.QueryInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						N: aws.String(strconv.Itoa(int(userID))),
					},
				},
			},
			"FileUniqueID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(sticker.FileUniqueID),
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
				N: aws.String(strconv.Itoa(int(sticker.UserID))),
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
