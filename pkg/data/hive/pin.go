package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/impartwealthapp/backend/pkg/data"
)

func (d *dynamo) SetPinStatus(hiveID, postID string, pin bool) error {
	updateInput := data.UpdateItemPropertyInput{
		DynamoDBAPI: d,
		TableName:   d.getTableNameForEnvironment(postTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"hiveId": {
				S: aws.String(hiveID),
			},
			"postId": {
				S: aws.String(postID),
			},
		},
		Update: expression.Set(expression.Name("isPinnedPost"), expression.Value(pin)),
		Logger: d.Logger,
	}

	return data.UpdateItemProperty(updateInput)
}

func (d *dynamo) PinPost(hiveID, postID string) error {
	updateInput := data.UpdateItemPropertyInput{
		DynamoDBAPI: d,
		TableName:   d.getTableNameForEnvironment(hiveTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"hiveId": {
				S: aws.String(hiveID),
			},
		},
		Update: expression.Set(expression.Name("pinnedPostId"), expression.Value(postID)),
		Logger: d.Logger,
	}

	return data.UpdateItemProperty(updateInput)
}
