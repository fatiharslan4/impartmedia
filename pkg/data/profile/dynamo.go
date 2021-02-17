package profile

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/impartwealthapp/backend/pkg/data"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/leebenson/conform"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const profileTableName = "profile"
const hiveTableName = "hive"
const authenticationIdIndexName = "idx_authenticationId"
const emailIndexName = "idx_email"
const screenNameIndexName = "idx_screenName"

func (d *ProfileDynamoDb) getTableNameForEnvironment(tableName string) string {
	if d.tableEnvironment != "" {
		return fmt.Sprintf("%s_%s", d.tableEnvironment, tableName)
	} else {
		return tableName
	}
}

type ProfileDynamoDb struct {
	Logger *zap.SugaredLogger
	dynamodbiface.DynamoDBAPI
	tableEnvironment string
}

func New(region, endpoint, environment string, logger *zap.SugaredLogger) (Store, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	var svc dynamodbiface.DynamoDBAPI
	if endpoint != "" {
		svc = dynamodb.New(sess, &aws.Config{Endpoint: aws.String(endpoint)})
	} else {
		svc = dynamodb.New(sess)
	}

	profileStore := &ProfileDynamoDb{
		DynamoDBAPI:      svc,
		Logger:           logger,
		tableEnvironment: environment,
	}

	if environment == "local" || environment == "dev" {
		resp, err := svc.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(profileStore.getTableNameForEnvironment(profileTableName))})
		if err != nil {
			return nil, errors.Wrap(err, "error retrieving info for "+profileStore.getTableNameForEnvironment(profileTableName))
		}
		if *resp.Table.TableName != profileStore.getTableNameForEnvironment(profileTableName) {
			return nil, errors.New(fmt.Sprintf("Expected table name of %s, got %s", profileStore.getTableNameForEnvironment(profileTableName), *resp.Table.TableName))
		}
	}

	return profileStore, nil
}

func (d *ProfileDynamoDb) getImpartId(attributeName, attributeValue, indexName string) (string, error) {
	d.Logger.Debug(fmt.Sprintf("looking up %s:%s using %s", attributeName, attributeValue, indexName))
	var impartId string

	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(attributeValue),
			},
		},
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :v1", attributeName)),
		ProjectionExpression:   aws.String("impartWealthId"),
		TableName:              aws.String(d.getTableNameForEnvironment(profileTableName)),
		IndexName:              aws.String(indexName),
	}

	resp, err := d.Query(input)
	if err != nil {
		return impartId, err
	}
	if *resp.Count == 0 {
		d.Logger.Debug("found no results", attributeName, attributeValue, indexName)
		return impartId, impart.ErrNotFound
	}

	if *resp.Count > 2 {
		d.Logger.Info(fmt.Sprintf("%s %s returned %v results!", attributeName, attributeValue, resp.Count))
		return impartId, impart.ErrUnknown
	}

	return *resp.Items[0]["impartWealthId"].S, nil
}

func (d *ProfileDynamoDb) GetImpartIdFromAuthId(authenticationId string) (string, error) {
	return d.getImpartId("authenticationId", authenticationId, authenticationIdIndexName)
}

func (d *ProfileDynamoDb) GetImpartIdFromEmail(email string) (string, error) {
	type T1 struct {
		Email string `conform:"email"`
	}
	t := T1{Email: email}
	conform.Strings(&t)
	return d.getImpartId("email", t.Email, emailIndexName)
}

func (d *ProfileDynamoDb) GetImpartIdFromScreenName(screenName string) (string, error) {
	return d.getImpartId("screenName", screenName, screenNameIndexName)
}

func (d *ProfileDynamoDb) CreateProfile(p models.Profile) (models.Profile, error) {
	out := models.Profile{}

	av, err := dynamodbattribute.MarshalMap(p)
	if err != nil {
		return out, err
	}

	input := &dynamodb.PutItemInput{
		Item:                   av,
		ReturnConsumedCapacity: aws.String("NONE"),
		TableName:              aws.String(d.getTableNameForEnvironment(profileTableName)),
		ConditionExpression:    aws.String("attribute_not_exists (impartWealthId)"),
		ReturnValues:           aws.String("NONE"),
	}

	_, err = d.PutItem(input)
	if err != nil {
		return out, err
	}

	return d.GetProfile(p.ImpartWealthID, true)

}

func (d *ProfileDynamoDb) GetProfileFromAuthId(authenticationId string, consistentRead bool) (models.Profile, error) {
	out := models.Profile{}

	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(authenticationId),
			},
		},
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :v1", "authenticationId")),
		TableName:              aws.String(d.getTableNameForEnvironment(profileTableName)),
		IndexName:              aws.String(authenticationIdIndexName),
	}

	resp, err := d.Query(input)
	if err != nil {
		return out, err
	}

	if resp.Items == nil {
		d.Logger.Debug("query return nil", resp)
		return out, impart.ErrNotFound
	}

	if len(resp.Items) == 0 {
		d.Logger.Desugar().Error("Unable to find matching profile", zap.String("authenticationId", authenticationId))
		return out, impart.ErrNotFound
	}

	if len(resp.Items) > 1 {
		d.Logger.Desugar().Error("DB Bad State! Critical error of 2 profiles for a given authentication ID user.", zap.String("authenticationId", authenticationId))
		return out, impart.ErrUnknown
	}

	err = dynamodbattribute.UnmarshalMap(resp.Items[0], &out)
	if err != nil {
		d.Logger.Error(err)
		return out, err
	}

	return out, nil
}

func (d *ProfileDynamoDb) GetProfile(impartWealthId string, consistentRead bool) (models.Profile, error) {
	out := models.Profile{}

	d.Logger.Debug(fmt.Sprintf("looking up impartWealthId: %s", impartWealthId))

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"impartWealthId": {
				S: aws.String(impartWealthId),
			},
		},
		TableName:      aws.String(d.getTableNameForEnvironment(profileTableName)),
		ConsistentRead: aws.Bool(consistentRead),
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

func (d *ProfileDynamoDb) UpdateProfile(authId string, p models.Profile) (models.Profile, error) {
	out := models.Profile{}

	av, err := dynamodbattribute.MarshalMap(p)
	if err != nil {
		return out, err
	}

	existingEntry := expression.Name("impartWealthId").AttributeExists()
	//authIdMatches := expression.Name("authenticationId").Equal(expression.Value(authId))
	//conditions := expression.And(existingEntry, authIdMatches)
	expr, err := expression.NewBuilder().WithCondition(existingEntry).Build()

	input := &dynamodb.PutItemInput{
		Item:                      av,
		ReturnConsumedCapacity:    aws.String("NONE"),
		TableName:                 aws.String(d.getTableNameForEnvironment(profileTableName)),
		ConditionExpression:       expr.Condition(), //aws.String("attribute_exists (impartWealthId)"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = d.PutItem(input)
	if err != nil {
		return out, expectedUpdateErr(err)
	}
	return d.GetProfile(p.ImpartWealthID, true)
}

func (d *ProfileDynamoDb) UpdateProfileProperty(impartWealthID, propertyPathName string, propertyPathValue interface{}) error {
	return d.updateProperty(profileTableName, impartWealthID, propertyPathName, propertyPathValue)
}

func (d *ProfileDynamoDb) UpdateWhiteListProperty(impartWealthID, propertyPathName string, propertyPathValue interface{}) error {
	return d.updateProperty(whitelistTableName, impartWealthID, propertyPathName, propertyPathValue)
}

func (d *ProfileDynamoDb) updateProperty(tableName, impartWealthID, propertyPathName string, propertyPathValue interface{}) error {

	updateInput := data.UpdateItemPropertyInput{
		DynamoDBAPI: d,
		TableName:   d.getTableNameForEnvironment(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"impartWealthId": {
				S: aws.String(impartWealthID),
			},
		},
		Update: expression.Set(expression.Name(propertyPathName), expression.Value(propertyPathValue)),
		Logger: d.Logger.Desugar(),
	}

	return data.UpdateItemProperty(updateInput)
}

func (d *ProfileDynamoDb) DeleteProfile(impartWealthID string) error {

	d.Logger.Debug(fmt.Sprintf("deleting impartWealthId: %s", impartWealthID))

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"impartWealthId": {
				S: aws.String(impartWealthID),
			},
		},
		TableName: aws.String(d.getTableNameForEnvironment(profileTableName)),
	}

	_, err := d.DeleteItem(input)

	return d.handleAWSErr(err)
}

func (d *ProfileDynamoDb) getProfiles(expr *expression.Expression, nextPage *models.NextProfilePage) ([]models.Profile, *models.NextProfilePage, error) {
	out := make([]models.Profile, 0)
	var err error

	scan := dynamodb.ScanInput{
		TableName: aws.String(d.getTableNameForEnvironment(profileTableName)),
	}

	if expr != nil {
		scan.FilterExpression = expr.Condition()

		if len(expr.Names()) > 0 {
			scan.ExpressionAttributeNames = expr.Names()
		}

		if len(expr.Values()) > 0 {
			scan.ExpressionAttributeValues = expr.Values()
		}
	}

	if nextPage != nil {
		scan.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"impartWealthId": {
				S: aws.String(nextPage.ImpartWealthID),
			},
		}
	}

	resp, err := d.Scan(&scan)
	if err != nil {
		d.Logger.Error("error scanning dynamo for profiles with a valid device token")
		return out, nil, err
	}

	d.Logger.Info("profile table scan complete", zap.Int64("scanCount", *resp.ScannedCount),
		zap.Int64("resultCount", *resp.Count))

	if *resp.Count == 0 {
		d.Logger.Info("no valid profiles returned")
		return out, nil, nil
	}

	nextPage = nil

	if len(resp.LastEvaluatedKey) > 0 {
		nextPage = &models.NextProfilePage{}
		if err = dynamodbattribute.UnmarshalMap(resp.LastEvaluatedKey, nextPage); err != nil {
			return out, nil, err
		}
	} else {
		nextPage = nil
	}

	if err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &out); err != nil {
		return out, nil, err
	}
	return out, nextPage, nil
}

func (d *ProfileDynamoDb) GetNotificationProfiles(nextPage *models.NextProfilePage) ([]models.Profile, *models.NextProfilePage, error) {
	out := make([]models.Profile, 0)
	var err error

	cb, err := expression.NewBuilder().
		WithCondition(expression.Name("notificationProfile.deviceToken").
			AttributeExists()).Build()
	if err != nil {
		return out, nil, err
	}

	return d.getProfiles(&cb, nextPage)
}

func (d *ProfileDynamoDb) GetHive(hiveID string, consistentRead bool) (models.Hive, error) {
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
		return models.Hive{}, d.handleAWSErr(err)
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
	return out, nil
}

func (d *ProfileDynamoDb) handleAWSErr(err error) error {
	if err == nil {
		return nil
	}

	d.Logger.Error("Error talking to DynamoDB: ", err.Error())

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeResourceNotFoundException:
			return impart.ErrNotFound
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			d.Logger.Error("Provisioned throughput exceeded! ", err.Error())
		default:
			return err
		}
	}
	return err
}

func expectedUpdateErr(err error) error {
	if err == nil {
		return nil
	}

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeConditionalCheckFailedException:
			return impart.ErrNotFound
		default:
			return err
		}
	}
	return err
}
