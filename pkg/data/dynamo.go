package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"go.uber.org/zap"
)

// UpdateItemPropertyInput is the input for UpdateItemProperty
type UpdateItemPropertyInput struct {
	dynamodbiface.DynamoDBAPI
	TableName string
	Update    expression.UpdateBuilder
	Key       map[string]*dynamodb.AttributeValue
	Logger    *zap.Logger
	Condition *expression.ConditionBuilder
}

// UpdateItemProperty Updates the item to the value from UpdateItemPropertyInput
func UpdateItemProperty(i UpdateItemPropertyInput) error {

	exprBuilder := expression.NewBuilder().WithUpdate(i.Update)

	if i.Condition != nil {
		exprBuilder.WithCondition(*i.Condition)
	}

	expr, err := exprBuilder.Build()
	if err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		Key:                       i.Key,
		TableName:                 aws.String(i.TableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	}

	if i.Condition != nil {
		input.ConditionExpression = expr.Condition()
	}

	out, err := i.UpdateItem(input)
	i.Logger.Debug("successfully updated item", zap.Any("awsUpdate", out))
	return err
}
