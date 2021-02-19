package data

import (
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

const postTableName = "hive_post"
const lsiPostTableDateTime = "lsi_postDatetime"
const lsiPostCommentDatetime = "lsi_lastCommentDatetime"
const gsiPostTableImpartID = "gsi_impartWealthId"

// Posts is the interface for Hive Post CRUD operations
type Posts interface {
	GetPosts(getPostsInput GetPostsInput) (models.Posts, *models.NextPage, error)
	GetPostsImpartWealthID(impartWealthID string, limit int64, offset time.Time) (models.Posts, error)
	GetPost(hiveID, postID string, consistentRead bool) (models.Post, error)
	SetPinStatus(hiveID, postID string, pin bool) error

	NewPost(post models.Post) (models.Post, error)
	EditPost(post models.Post) (models.Post, error)
	IncrementDecrementPost(hiveID, postID, colName string, subtract bool) error
	UpdateTimestampLater(hiveID, postID, colName string, value time.Time) error

	DeletePost(hiveID, postID string) error
}

// NewPostData returns an implementation of data.Posts interface
func NewPostData(region, endpoint, environment string, logger *zap.Logger) (Posts, error) {
	return newDynamo(region, endpoint, environment, logger)
}

func (d *dynamo) IncrementDecrementPost(hiveID, postID, colName string, subtract bool) error {
	return IncrementDecrement(IncrementDecrementInput{
		DynamoDBAPI: d.DynamoDBAPI,
		TableName:   d.getTableNameForEnvironment(postTableName),
		ColumnName:  colName,
		Subtract:    subtract,
		Logger:      d.Logger,
		Key:         postKey(hiveID, postID),
	})
}

func (d *dynamo) UpdateTimestampLater(hiveID, postID, colName string, value time.Time) error {
	condition := expression.Name(colName).LessThan(expression.Value(value))
	return UpdateItemProperty(UpdateItemPropertyInput{
		DynamoDBAPI: d.DynamoDBAPI,
		TableName:   d.getTableNameForEnvironment(postTableName),
		Update:      expression.Set(expression.Name(colName), expression.Value(value)),
		Logger:      d.Logger,
		Key:         postKey(hiveID, postID),
		Condition:   &condition,
	})
}

func postKey(hiveID, postID string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"hiveId": {
			S: aws.String(hiveID),
		},
		"postId": {
			S: aws.String(postID),
		},
	}
}

// GetPost gets a single post and it's associated content from dynamodb.
func (d *dynamo) GetPost(hiveID, postID string, consistentRead bool) (models.Post, error) {

	input := &dynamodb.GetItemInput{
		Key:            postKey(hiveID, postID),
		TableName:      aws.String(d.getTableNameForEnvironment(postTableName)),
		ConsistentRead: aws.Bool(consistentRead),
	}

	resp, err := d.GetItem(input)
	if err != nil {
		d.Logger.Error("error getting item from dynamodb", zap.Error(err))
		return models.Post{}, handleAWSErr(err)
	}

	if resp.Item == nil {
		d.Logger.Debug("get item return null", zap.Error(err))
		return models.Post{}, impart.ErrNotFound
	}

	var out models.Post
	err = dynamodbattribute.UnmarshalMap(resp.Item, &out)
	if err != nil {
		d.Logger.Error("Error trying to unmarshal attribute", zap.Error(err))
		return models.Post{}, err
	}
	d.Logger.Debug("retrieved", zap.Any("post", out))

	return out, nil
}

// NewPost Creates a new Post in DynamoDB
func (d *dynamo) NewPost(post models.Post) (models.Post, error) {
	var err error

	d.Logger.Debug("Creating new")
	item, err := dynamodbattribute.MarshalMap(post)
	if err != nil {
		return models.Post{}, err
	}

	cb := expression.Name("hiveId").AttributeNotExists().And(expression.Name("postId").AttributeNotExists())
	expr, err := expression.NewBuilder().
		WithCondition(cb).
		Build()
	if err != nil {
		return models.Post{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:                     item,
		ReturnConsumedCapacity:   aws.String("NONE"),
		TableName:                aws.String(d.getTableNameForEnvironment(postTableName)),
		ConditionExpression:      expr.Condition(), // aws.String("attribute_not_exists(hiveId)"),
		ExpressionAttributeNames: expr.Names(),
		//ExpressionAttributeValues: expr.Values(),
		ReturnValues: aws.String("NONE"),
	}

	_, err = d.PutItem(input) //2019-01-24T23:12:37.601353-08:00
	if err != nil {
		return models.Post{}, err
	}

	return d.GetPost(post.HiveID, post.PostID, true)
}

// EditPost takes an incoming Post, and modifies the DynamoDB record to match.
func (d *dynamo) EditPost(post models.Post) (models.Post, error) {
	var err error

	item, err := dynamodbattribute.MarshalMap(post)
	if err != nil {
		return models.Post{}, err
	}

	if len(post.Edits) > 1 {
		post.Edits.SortDescending()
	}

	cb := expression.Name("hiveId").AttributeExists().And(expression.Name("postId").AttributeExists())
	expr, err := expression.NewBuilder().
		WithCondition(cb).
		Build()
	if err != nil {
		return models.Post{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:                     item,
		ReturnConsumedCapacity:   aws.String("NONE"),
		TableName:                aws.String(d.getTableNameForEnvironment(postTableName)),
		ConditionExpression:      expr.Condition(),
		ExpressionAttributeNames: expr.Names(),
		ReturnValues:             aws.String("NONE"),
	}

	_, err = d.PutItem(input)
	if err != nil {
		d.Logger.Error("Error replacing dynamoDB item", zap.Error(err), zap.Any("hive", post))
		return models.Post{}, err
	}

	return d.GetPost(post.HiveID, post.PostID, true)
}

// GetPostsInput is the input necessary
type GetPostsInput struct {
	// HiveID is the ID that should be queried for posts
	HiveID string
	// Limit is the maximum number of records that should be returns.  The API can optionally return
	// less than Limit, if DynamoDB decides the items read were too large.
	Limit int64
	// Offset is the optional is for getting the next page if not all results were included.
	NextPage *models.NextPage
	// IsLastCommentSorted Changes the sort from default of PostDatetime to LastCommentDatetime
	// Default: false
	IsLastCommentSorted bool
	// Tags is the optional list of tags to filter on
	TagIDs []int
}

type commentSortedPageKey struct {
	HiveID              string    `json:"hiveId"`
	PostID              string    `json:"postId"`
	LastCommentDatetime time.Time `json:"lastCommentDatetime"`
}

type postSortedPageKey struct {
	HiveID       string    `json:"hiveId"`
	PostID       string    `json:"postId"`
	PostDatetime time.Time `json:"postDatetime"`
}

// GetPosts takes a set GetPostsInput, and decides based on this input how to query DynamoDB.
func (d *dynamo) GetPosts(gpi GetPostsInput) (models.Posts, *models.NextPage, error) {
	out := make(models.Posts, 0)
	var offset *models.NextPage
	var err error
	var requestedLimit = gpi.Limit

	kc := expression.Key("hiveId").
		Equal(expression.Value(gpi.HiveID))

	exprBuilder := expression.NewBuilder().WithKeyCondition(kc)
	dynamoRequestSize := DefaultLimit

	var f expression.ConditionBuilder
	if len(gpi.TagIDs) > 0 {
		for i := 0; i < len(gpi.TagIDs); i++ {
			if i == 0 {
				f = expression.Name("tags").Contains(strconv.Itoa(gpi.TagIDs[i]))
			}
			f.Or(expression.Name("tags").Contains(strconv.Itoa(gpi.TagIDs[i])))
		}
		exprBuilder.WithFilter(f)
	}
	//set requested page size if unset, set it to the default.
	if requestedLimit <= 0 {
		requestedLimit = DefaultLimit
	}

	//if the requested page size is less that the default, set the dynamo request size to the default to be more efficient.
	if requestedLimit <= DefaultLimit {
		dynamoRequestSize = DefaultLimit
	} else {
		dynamoRequestSize = requestedLimit
	}

	expr, err := exprBuilder.Build()
	if err != nil {
		d.Logger.Error("Error building DynamoDB Condition", zap.Error(err))
		return out, offset, err
	}
	offset = gpi.NextPage
	getNextPage := true

	//fix expression attribute valutes to move tag strings to int
	values := expr.Values()
	for k, v := range expr.Names() {
		if *v == "tags" {
			k = strings.Replace(k, "#", ":", 1)
			vv := values[k]
			vv.N = vv.S
			vv.S = nil
		}
	}
	for getNextPage {
		input := &dynamodb.QueryInput{
			TableName:                 aws.String(d.getTableNameForEnvironment(postTableName)),
			ConsistentRead:            aws.Bool(false),
			Limit:                     aws.Int64(dynamoRequestSize),
			ScanIndexForward:          aws.Bool(false),
			KeyConditionExpression:    expr.KeyCondition(),
			FilterExpression:          expr.Filter(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		}

		err = addIndexSortAttributes(input, gpi, offset)
		if err != nil {
			return out, offset, err
		}

		d.Logger.Debug("DynamoDB Query", zap.Any("input", input))

		resp, err := d.Query(input)
		if err != nil {
			d.Logger.Error("error querying posts",
				zap.Any("input", gpi),
				zap.Error(err))
			return out, offset, handleAWSErr(err)
		}

		if resp.Items == nil {
			d.Logger.Debug("get items returned nil",
				zap.Any("input", gpi),
				zap.Error(err))
			return out, offset, impart.ErrNotFound
		}

		d.Logger.Debug("retrieved",
			zap.Any("input", gpi),
			zap.Any("query", input),
			zap.Any("resp", resp))

		page := make(models.Posts, 0)
		err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &page)
		if err != nil {
			d.Logger.Error("Error trying to unmarshal posts",
				zap.Any("input", gpi),
				zap.Error(err))
			return out, offset, err
		}

		out = append(out, page...)
		out.SortDescending(gpi.IsLastCommentSorted)

		offset, err = handleNextPage(gpi.IsLastCommentSorted, resp)
		if err != nil {
			return out, offset, err
		}

		if len(out) > int(requestedLimit) {
			d.Logger.Debug("requested limit exceeded - limiting result set", zap.Int64("limit", requestedLimit))
			out = out[:requestedLimit]

			offset = &models.NextPage{
				ContentID: out[requestedLimit-1].PostID,
			}
			if gpi.IsLastCommentSorted {
				offset.Timestamp = out[requestedLimit-1].LastCommentDatetime
			} else {
				offset.Timestamp = out[requestedLimit-1].PostDatetime
			}

			getNextPage = false
		} else if len(out) == int(requestedLimit) {
			getNextPage = false
		} else if offset == nil {
			getNextPage = false
		}
	}

	d.Logger.Debug("dynamo posts request response", zap.Any("input", gpi), zap.Any("posts", out))
	return out, offset, nil
}

func handleNextPage(isLastCommentSorted bool, resp *dynamodb.QueryOutput) (*models.NextPage, error) {
	var nextPage *models.NextPage
	var err error

	if len(resp.LastEvaluatedKey) <= 0 {
		return nextPage, err
	}

	if isLastCommentSorted {
		var lastKey commentSortedPageKey
		if err = dynamodbattribute.UnmarshalMap(resp.LastEvaluatedKey, &lastKey); err != nil {
			return nil, err
		}
		nextPage = &models.NextPage{
			ContentID: lastKey.PostID,
			Timestamp: lastKey.LastCommentDatetime,
		}

	} else {
		var lastKey postSortedPageKey
		if err = dynamodbattribute.UnmarshalMap(resp.LastEvaluatedKey, &lastKey); err != nil {
			return nil, err
		}
		nextPage = &models.NextPage{
			ContentID: lastKey.PostID,
			Timestamp: lastKey.PostDatetime,
		}
	}

	return nextPage, err
}

func addIndexSortAttributes(input *dynamodb.QueryInput, gpi GetPostsInput, offset *models.NextPage) error {

	var sortAttributes map[string]*dynamodb.AttributeValue
	var err error

	if gpi.IsLastCommentSorted {
		input.IndexName = aws.String(lsiPostCommentDatetime)
		if offset != nil {
			sortAttributes, err = dynamodbattribute.MarshalMap(commentSortedPageKey{
				HiveID: gpi.HiveID, PostID: offset.ContentID, LastCommentDatetime: offset.Timestamp,
			})
			if err != nil {
				return err
			}
			input.SetExclusiveStartKey(sortAttributes)
		}
	} else {
		input.IndexName = aws.String(lsiPostTableDateTime)
		if offset != nil {
			sortAttributes, err = dynamodbattribute.MarshalMap(postSortedPageKey{
				HiveID: gpi.HiveID, PostID: offset.ContentID, PostDatetime: offset.Timestamp,
			})
			if err != nil {
				return err
			}
			input.SetExclusiveStartKey(sortAttributes)
		}
	}

	return nil
}

// GetPostsImpartWealthID takes an inputImpartWealthID and optional limit and offset and returns
// a list of posts for that impartWealthId
func (d *dynamo) GetPostsImpartWealthID(impartWealthID string, limit int64, offset time.Time) (models.Posts, error) {
	out := make(models.Posts, 0)
	var err error

	kc := expression.Key("impartWealthId").Equal(expression.Value(impartWealthID))

	expr, err := expression.NewBuilder().WithKeyCondition(kc).Build()
	if err != nil {
		d.Logger.Error("Error building DynamoDB Condition", zap.Error(err))
		return out, err
	}

	if limit <= 0 {
		limit = DefaultLimit
	}
	input := &dynamodb.QueryInput{
		TableName:                 aws.String(d.getTableNameForEnvironment(postTableName)),
		ConsistentRead:            aws.Bool(false),
		Limit:                     aws.Int64(limit),
		ScanIndexForward:          aws.Bool(false),
		IndexName:                 aws.String(gsiPostTableImpartID),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	resp, err := d.Query(input)
	if err != nil {
		d.Logger.Error("error querying for posts",
			zap.String("impartWealthId", impartWealthID),
			zap.Error(err))
		return out, handleAWSErr(err)
	}

	if resp.Items == nil {
		d.Logger.Error("get items returned nil",
			zap.String("impartWealthId", impartWealthID),
			zap.Error(err))
		return out, impart.ErrNotFound
	}

	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &out)
	if err != nil {
		d.Logger.Error("error unmarshalling dynamodb attributs",
			zap.String("impartWealthId", impartWealthID),
			zap.Error(err))
		return out, err
	}

	d.Logger.Debug("retrieved", zap.Any("posts", out))

	return out, nil
}

func (d *dynamo) DeletePost(hiveID, postID string) error {
	err := d.DeleteCommentsBatch(postID)
	if err != nil {
		return err
	}

	err = d.DeleteTracks(postID)
	if err != nil {
		return err
	}

	_, err = d.DynamoDBAPI.DeleteItem(&dynamodb.DeleteItemInput{
		Key:       postKey(hiveID, postID),
		TableName: aws.String(d.getTableNameForEnvironment(postTableName)),
	})
	return handleAWSErr(err)
}
