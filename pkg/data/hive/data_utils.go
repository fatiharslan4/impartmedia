package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"go.uber.org/zap"
)

// IncrementDecrementInput is the input for IncrementDecrementPost
type IncrementDecrementInput struct {
	dynamodbiface.DynamoDBAPI
	TableName  string
	ColumnName string
	Key        map[string]*dynamodb.AttributeValue
	Subtract   bool
	Logger     *zap.Logger
}

// IncrementDecrement increments the column by 1 or -1 depending on the subtract field of IncrementDecrementInput
func IncrementDecrement(i IncrementDecrementInput) error {
	var increment = 1
	if i.Subtract {
		increment = -1
	}

	input := &dynamodb.UpdateItemInput{
		Key:       i.Key,
		TableName: aws.String(i.TableName),
	}

	upd := expression.Set(expression.Name(i.ColumnName), expression.Name(i.ColumnName).Plus(expression.Value(increment)))
	exprBuilder := expression.NewBuilder().WithUpdate(upd)
	if i.Subtract {
		cb := expression.Name(i.ColumnName).GreaterThan(expression.Value(0))
		exprBuilder.WithCondition(cb)
	}

	expr, err := exprBuilder.Build()
	if err != nil {
		return err
	}

	input.ExpressionAttributeValues = expr.Values()
	input.ExpressionAttributeNames = expr.Names()
	input.UpdateExpression = expr.Update()
	if i.Subtract {
		input.ConditionExpression = expr.Condition()
	}

	out, err := i.UpdateItem(input)
	if conditionalUpdateNoError(err) != nil {
		return err
	}

	i.Logger.Debug("successfully updated item", zap.Any("awsUpdate", out))
	return nil
}

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

//
//// GetSortKeys is the input necessary to get a set of sort keys by their key and name
//type GetSortKeys struct {
//	TableName string
//	// KeyValue is the ID that should be queried for
//	KeyValue string
//	// KeyName is the
//	KeyName string
//	// AscendingSort should be true if you want to traverse the sort key by ascending order
//	AscendingSort bool
//	// Limit is the maximum number of records that should be returns.  The API can optionally return
//	// less than Limit, if DynamoDB decides the items read were too large.
//	Limit int64
//	// Offset is the optional is for getting the next page if not all results were included.
//	Offset time.Time
//	// IsLastCommentSorted Changes the sort from default of PostDatetime to LastCommentDatetime
//	// Default: false
//	IsLastCommentSorted bool
//
//	// Filter is the dynamoDB expression filter to apply
//	Filter *expression.ConditionBuilder
//
//	// SortAttributes is the expected sort attributes for paging on the sort keys
//	SortAttributes map[string]*dynamodb.AttributeValue
//
//	// IndexName is the index to use to scan
//	IndexName string
//}
//
//// GetPosts takes a set GetPostsInput, and decides based on this input how to query DynamoDB.
//func (d *dynamo) GetItems(gpi GetSortKeys) (interface{}, error) {
//	out := make(models.Posts, 0)
//	var err error
//
//	kc := expression.Key(gpi.KeyName).
//		Equal(expression.Value(gpi.KeyValue))
//
//	exprBuilder := expression.NewBuilder().WithKeyCondition(kc)
//
//	var f expression.ConditionBuilder
//	if gpi.Filter != nil {
//		exprBuilder.WithFilter(f)
//	}
//
//	expr, err := exprBuilder.Build()
//	if err != nil {
//		d.Logger.Error("Error building DynamoDB Condition", zap.Error(err))
//		return out, err
//	}
//
//	if gpi.Limit <= 0 {
//		gpi.Limit = DefaultLimit
//	}
//
//	input := &dynamodb.QueryInput{
//		TableName:                 aws.String(d.getTableNameForEnvironment(gpi.TableName)),
//		ConsistentRead:            aws.Bool(false),
//		Limit:                     aws.Int64(gpi.Limit),
//		ScanIndexForward:          aws.Bool(gpi.AscendingSort),
//		KeyConditionExpression:    expr.KeyCondition(),
//		FilterExpression:          expr.Filter(),
//		ExpressionAttributeNames:  expr.Names(),
//		ExpressionAttributeValues: expr.Values(),
//		IndexName:                 aws.String(gpi.IndexName),
//	}
//
//	if !gpi.Offset.Equal(DefaultTime) {
//		input.SetExclusiveStartKey(gpi.SortAttributes)
//	}
//
//	resp, err := d.Query(input)
//	if err != nil {
//		d.Logger.Error("error querying posts",
//			zap.Any("input", gpi),
//			zap.Error(err))
//		return out, handleAWSErr(err)
//	}
//
//	if resp.Items == nil {
//		d.Logger.Debug("get items returned nil",
//			zap.Any("input", gpi),
//			zap.Error(err))
//		return out, impart.ErrNotFound
//	}
//
//	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &out)
//	if err != nil {
//		d.Logger.Error("Error trying to unmarshal posts",
//			zap.Any("input", gpi),
//			zap.Error(err))
//		return out, err
//	}
//
//	d.Logger.Debug("retrieved",
//		zap.Any("input", gpi),
//		zap.Any("posts", out))
//
//	return out, nil
//}
