package stickers

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// TODO: Optimize session usage

// GetStickers gets a list of stickers by user id
func GetStickers(userID int) ([]Sticker, error) {
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

// Put inserts a sticker into table
func (sticker Sticker) Put() error {
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

// BelongsToUser checks if user has a sticker
func (sticker Sticker) BelongsToUser(userID int) (bool, error) {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})

	client := dynamodb.New(sess)

	result, err := client.Query(&dynamodb.QueryInput{
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

// Delete removes a sticker
func (sticker Sticker) Delete() error {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})

	client := dynamodb.New(sess)

	_, err := client.DeleteItem(&dynamodb.DeleteItemInput{
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
