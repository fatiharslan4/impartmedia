package profile

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

const whitelistTableName = "whitelist_profile"

func (d *ProfileDynamoDb) CreateWhitelistEntry(p models.WhiteListProfile) error {
	av, err := dynamodbattribute.MarshalMap(p)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:                   av,
		ReturnConsumedCapacity: aws.String("NONE"),
		TableName:              aws.String(d.getTableNameForEnvironment(whitelistTableName)),
		ReturnValues:           aws.String("NONE"),
	}

	_, err = d.PutItem(input)
	return err
}

func (d *ProfileDynamoDb) GetWhitelistEntry(impartWealthID string) (models.WhiteListProfile, error) {
	out := models.WhiteListProfile{}

	d.Logger.Debug(fmt.Sprintf("looking up impartWealthId: %s", impartWealthID))

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"impartWealthId": {
				S: aws.String(impartWealthID),
			},
		},
		TableName:      aws.String(d.getTableNameForEnvironment(whitelistTableName)),
		ConsistentRead: aws.Bool(true),
	}

	resp, err := d.GetItem(input)
	if err != nil {
		d.Logger.Error(err)
		return out, d.handleAWSErr(err)
	}

	if resp.Item == nil {
		d.Logger.Debug("get item return null", resp)
		return out, impart.ErrNotFound
	}

	err = dynamodbattribute.UnmarshalMap(resp.Item, &out)
	if err != nil {
		d.Logger.Error(err)
		return out, err
	}

	return out, nil
}

func (d *ProfileDynamoDb) SearchWhitelistEntry(t SearchType, value string) (models.WhiteListProfile, error) {
	out := models.WhiteListProfile{}

	kc := expression.Key(t.String()).Equal(expression.Value(value))
	expr, err := expression.NewBuilder().WithKeyCondition(kc).Build()
	if err != nil {
		return out, err
	}

	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.getTableNameForEnvironment(whitelistTableName)),
		IndexName:                 aws.String(t.IndexName()),
	}

	resp, err := d.Query(input)
	if err != nil {
		d.Logger.Debug("dynamo response had an error", zap.Error(err), zap.Any("response", resp))
		return out, d.handleAWSErr(err)
	}

	if resp.Items == nil {
		d.Logger.Desugar().Debug("query returned nil", zap.Any("response", resp), zap.String(t.String(), value))
		return out, impart.ErrNotFound
	}

	if len(resp.Items) == 0 {
		return out, impart.ErrNotFound
	}

	if len(resp.Items) > 1 {
		d.Logger.Desugar().Error("DB Bad State! Critical error of 2 whitelist entries for a given email", zap.String(t.String(), value))
		return out, impart.ErrUnknown
	}

	err = dynamodbattribute.UnmarshalMap(resp.Items[0], &out)
	if err != nil {
		d.Logger.Error(err)
		return out, err
	}

	return d.GetWhitelistEntry(out.ImpartWealthID)
}

func (d *ProfileDynamoDb) UpdateWhitelistEntryScreenName(impartWealthID, screenName string) error {
	return d.UpdateWhiteListProperty(impartWealthID, "screenName", screenName)
}
