package data

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type dynamo struct {
	Logger *zap.Logger
	dynamodbiface.DynamoDBAPI
	tableEnvironment string
}

const hiveTableName = "hive"

// Hives is the interface for Hive CRUD operations
type Hives interface {
	GetHives() (models.Hives, error)
	GetHive(hiveID string, consistentRead bool) (models.Hive, error)
	NewHive(hive models.Hive) (models.Hive, error)
	EditHive(hive models.Hive) (models.Hive, error)
	PinPost(hiveID, PostID string) error
}

func newDynamo(region, endpoint, environment string, logger *zap.Logger) (*dynamo, error) {
	h := &dynamo{
		tableEnvironment: environment,
		Logger:           logger,
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	if endpoint != "" {
		h.DynamoDBAPI = dynamodb.New(sess,
			&aws.Config{
				Endpoint:   aws.String(endpoint),
				HTTPClient: impart.ImpartHttpClient(30 * time.Second),
			})
	} else {
		h.DynamoDBAPI = dynamodb.New(sess)
	}

	if environment == "local" || environment == "dev" {
		tableName := h.getTableNameForEnvironment(hiveTableName)
		resp, err := h.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
		if err != nil {
			return nil, errors.Wrap(err, "error retrieving info for "+tableName)
		}
		if *resp.Table.TableName != tableName {
			return nil, fmt.Errorf("expected table name of %s, got %s", tableName, *resp.Table.TableName)
		}
	}
	return h, nil
}

// NewHiveData returns an implementation of data.Hives
func NewHiveData(region, endpoint, environment string, logger *zap.Logger) (Hives, error) {
	return newDynamo(region, endpoint, environment, logger)
}

func (d *dynamo) GetHives() (models.Hives, error) {
	out := make(models.Hives, 0)
	var err error

	input := &dynamodb.ScanInput{
		TableName:      aws.String(d.getTableNameForEnvironment(hiveTableName)),
		ConsistentRead: aws.Bool(false),
		Limit:          aws.Int64(100),
	}

	resp, err := d.Scan(input)
	if err != nil {
		d.Logger.Error("error scanning for hives in dynamodb", zap.Error(err))
		return out, handleAWSErr(err)
	}

	if resp.Items == nil {
		d.Logger.Debug("get items returned nil", zap.Error(err))
		return out, impart.ErrNotFound
	}

	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &out)
	if err != nil {
		d.Logger.Error("Error trying to unmarshal list of hives", zap.Error(err))
		return out, err
	}

	d.Logger.Debug("retrieved", zap.Any("hive", out))
	return out, nil
}

func (d *dynamo) getTableNameForEnvironment(tableName string) string {
	if d.tableEnvironment != "" {
		return fmt.Sprintf("%s_%s", d.tableEnvironment, tableName)
	}
	return tableName
}

func (d *dynamo) GetHive(hiveID string, consistentRead bool) (models.Hive, error) {
	var out models.Hive
	var err error

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"hiveId": {
				S: aws.String(hiveID),
			},
		},
		TableName:      aws.String(d.getTableNameForEnvironment(hiveTableName)),
		ConsistentRead: aws.Bool(consistentRead),
	}

	resp, err := d.GetItem(input)
	if err != nil {
		d.Logger.Error("error getting item from dynamodb", zap.Error(err))
		return models.Hive{}, handleAWSErr(err)
	}

	if resp.Item == nil {
		d.Logger.Debug("get item return null", zap.Error(err))
		return models.Hive{}, impart.ErrNotFound
	}

	err = dynamodbattribute.UnmarshalMap(resp.Item, &out)
	if err != nil {
		d.Logger.Error("Error trying to unmarshal attribute", zap.Error(err))
		return models.Hive{}, err
	}
	d.Logger.Debug("retrieved", zap.Any("hive", out))

	out.HiveDistributions.CleanEmptyValues()
	if !out.HiveDistributions.IsSorted() {
		out.HiveDistributions.Sort()
	}

	return out, nil
}

func (d *dynamo) NewHive(hive models.Hive) (models.Hive, error) {
	var err error

	hive.HiveDistributions.CleanEmptyValues()
	if !hive.HiveDistributions.IsSorted() {
		hive.HiveDistributions.Sort()
	}

	item, err := dynamodbattribute.MarshalMap(hive)
	if err != nil {
		return models.Hive{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:                   item,
		ReturnConsumedCapacity: aws.String("NONE"),
		TableName:              aws.String(d.getTableNameForEnvironment(hiveTableName)),
		ConditionExpression:    aws.String("attribute_not_exists(hiveId)"),
		ReturnValues:           aws.String("NONE"),
	}

	_, err = d.PutItem(input)
	if err != nil {
		return models.Hive{}, err
	}

	return d.GetHive(hive.HiveID, true)
}

func (d *dynamo) EditHive(hive models.Hive) (models.Hive, error) {
	var err error

	hive.HiveDistributions.CleanEmptyValues()
	if !hive.HiveDistributions.IsSorted() {
		hive.HiveDistributions.Sort()
	}

	item, err := dynamodbattribute.MarshalMap(hive)
	if err != nil {
		return models.Hive{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:                   item,
		ReturnConsumedCapacity: aws.String("NONE"),
		TableName:              aws.String(d.getTableNameForEnvironment(hiveTableName)),
		ConditionExpression:    aws.String("attribute_exists (hiveId)"),
	}

	_, err = d.PutItem(input)
	if err != nil {
		d.Logger.Error("Error replacing dynamoDB item", zap.Error(err), zap.Any("hive", hive))
		return models.Hive{}, err
	}

	return d.GetHive(hive.HiveID, true)
}
