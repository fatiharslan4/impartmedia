package data

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/jpillora/backoff"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// UpVoteCountColumnName upVotes
const UpVoteCountColumnName = "upVotes"

// DownVoteCountColumnName downVotes
const DownVoteCountColumnName = "downVotes"

// CommentCountColumnName commentCount
const CommentCountColumnName = "commentCount"

// LastCommentDatetimeColumnName on models.Post
const LastCommentDatetimeColumnName = "lastCommentDatetime"

const commentTableName = "hive_comment"
const gsiCommentImpartID = "gsi_impartWealthId"
const lsiCommentDateTime = "lsi_commentDatetime"

// Comments is the interface for Hive Comment CRUD operations
type Comments interface {
	GetComments(postID string, limit int64, nextPage *models.NextPage) (models.Comments, *models.NextPage, error)
	GetCommentsByImpartWealthID(impartWealthID string, limit int64, offset time.Time) (models.Comments, error)

	GetComment(postID, commentID string, consistentRead bool) (models.Comment, error)

	NewComment(models.Comment) (models.Comment, error)
	EditComment(comment models.Comment) (models.Comment, error)

	IncrementDecrementComment(postID, commentID, colName string, subtract bool) error

	DeleteComment(postID, commentID string) error
	DeleteComments(postID string) error
}

// NewCommentData returns an implementation of data.Comments interface
func NewCommentData(region, endpoint, environment string, logger *zap.Logger) (Comments, error) {
	return newDynamo(region, endpoint, environment, logger)
}

// GetPost gets a single post and it's associated content from dynamodb.
func (d *dynamo) GetComment(postID, commentID string, consistentRead bool) (models.Comment, error) {

	input := &dynamodb.GetItemInput{
		Key:            commentKey(postID, commentID),
		TableName:      aws.String(d.getTableNameForEnvironment(commentTableName)),
		ConsistentRead: aws.Bool(consistentRead),
	}

	resp, err := d.GetItem(input)
	if err = handleAWSErr(err); err != nil {
		if err != impart.ErrNotFound {
			d.Logger.Error("error getting item from dynamodb", zap.Error(err))
		}
		return models.Comment{}, handleAWSErr(err)
	}

	if resp.Item == nil {
		d.Logger.Debug("get item return null", zap.Error(err))
		return models.Comment{}, impart.ErrNotFound
	}

	var out models.Comment
	err = dynamodbattribute.UnmarshalMap(resp.Item, &out)
	if err != nil {
		d.Logger.Error("Error trying to unmarshal attribute", zap.Error(err))
		return models.Comment{}, err
	}
	d.Logger.Debug("retrieved", zap.Any("comments", out))

	return out, nil
}

// NewComment Creates a new Comment for a comments in DynamoDB
func (d *dynamo) NewComment(comment models.Comment) (models.Comment, error) {
	var err error

	item, err := dynamodbattribute.MarshalMap(comment)
	if err != nil {
		return models.Comment{}, err
	}

	cb := expression.Name("postId").AttributeNotExists().
		And(expression.Name("commentId").AttributeNotExists())

	expr, err := expression.NewBuilder().
		WithCondition(cb).
		Build()
	if err != nil {
		return models.Comment{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:                     item,
		ReturnConsumedCapacity:   aws.String("NONE"),
		TableName:                aws.String(d.getTableNameForEnvironment(commentTableName)),
		ConditionExpression:      expr.Condition(),
		ExpressionAttributeNames: expr.Names(),
		//ExpressionAttributeValues: expr.Values(),
		ReturnValues: aws.String("NONE"),
	}

	_, err = d.PutItem(input)
	if err != nil {
		return models.Comment{}, err
	}

	return d.GetComment(comment.PostID, comment.CommentID, true)
}

// EditComment takes an incoming comments, and modifies the DynamoDB record to match.
func (d *dynamo) EditComment(comment models.Comment) (models.Comment, error) {
	var err error

	item, err := dynamodbattribute.MarshalMap(comment)
	if err != nil {
		return models.Comment{}, err
	}

	if len(comment.Edits) > 1 {
		comment.Edits.SortDescending()
	}

	cb := expression.Name("postId").AttributeExists().
		And(expression.Name("commentId").AttributeExists())
	expr, err := expression.NewBuilder().
		WithCondition(cb).
		Build()
	if err != nil {
		return models.Comment{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:                     item,
		ReturnConsumedCapacity:   aws.String("NONE"),
		TableName:                aws.String(d.getTableNameForEnvironment(commentTableName)),
		ConditionExpression:      expr.Condition(),
		ExpressionAttributeNames: expr.Names(),
		ReturnValues:             aws.String("NONE"),
	}

	_, err = d.PutItem(input)
	if err != nil {
		d.Logger.Error("Error replacing dynamoDB item", zap.Error(err), zap.Any("hive", comment))
		return models.Comment{}, err
	}

	return d.GetComment(comment.PostID, comment.CommentID, true)
}

// GetPosts takes a set GetPostsInput, and decides based on this input how to query DynamoDB.
func (d *dynamo) GetComments(postID string, limit int64, offset *models.NextPage) (models.Comments, *models.NextPage, error) {
	out := make(models.Comments, 0)
	var nextPage *models.NextPage
	var err error
	d.Logger.Debug("GetCommentsDataRequest", zap.String("postID", postID),
		zap.Int64("limit", limit),
		zap.Any("offset", offset))

	kc := expression.Key("postId").
		Equal(expression.Value(postID))

	exprBuilder := expression.NewBuilder().WithKeyCondition(kc)

	expr, err := exprBuilder.Build()
	if err != nil {
		d.Logger.Error("Error building DynamoDB Condition", zap.Error(err))
		return out, nextPage, err
	}

	if limit <= 0 {
		limit = DefaultLimit
	}

	type CommentPage struct {
		PostID          string    `json:"postId"`
		CommentID       string    `json:"commentId"`
		CommentDateTime time.Time `json:"commentDatetime"`
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(d.getTableNameForEnvironment(commentTableName)),
		ConsistentRead:            aws.Bool(false),
		IndexName:                 aws.String(lsiCommentDateTime),
		Limit:                     aws.Int64(limit),
		ScanIndexForward:          aws.Bool(true),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	if offset != nil {
		var sortAttributes map[string]*dynamodb.AttributeValue

		sortAttributes, err = dynamodbattribute.MarshalMap(CommentPage{
			PostID: postID, CommentDateTime: offset.Timestamp, CommentID: offset.ContentID,
		})
		if err != nil {
			return out, nextPage, err
		}

		d.Logger.Debug("Received non-default offset timestamp", zap.Any("commentSortedPageKey", sortAttributes))

		input.SetExclusiveStartKey(sortAttributes)
	}

	d.Logger.Debug("comments query", zap.Any("queryInput", input))

	resp, err := d.Query(input)
	if err != nil {
		d.Logger.Error("error querying comments",
			zap.String("postId", postID),
			zap.Int64("limit", limit),
			zap.Any("offset", offset),
			zap.Error(err))
		return out, nextPage, handleAWSErr(err)
	}

	if resp.Items == nil {
		d.Logger.Debug("get items returned nil",
			zap.String("postId", postID),
			zap.Int64("limit", limit),
			zap.Any("offset", offset),
			zap.Error(err))
		return out, nextPage, impart.ErrNotFound
	}

	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &out)
	if err != nil {
		d.Logger.Error("Error trying to unmarshal comments",
			zap.String("postId", postID),
			zap.Int64("limit", limit),
			zap.Any("offset", offset),
			zap.Error(err))
		return out, nextPage, err
	}

	if len(resp.LastEvaluatedKey) > 0 {
		lastKey := CommentPage{}
		err = dynamodbattribute.UnmarshalMap(resp.LastEvaluatedKey, &lastKey)
		nextPage = &models.NextPage{
			ContentID: lastKey.CommentID,
			Timestamp: lastKey.CommentDateTime,
		}
	}

	d.Logger.Debug("retrieved",
		zap.String("postId", postID),
		zap.Int64("limit", limit),
		zap.Any("offset", offset),
		zap.Any("nextPage", nextPage),
		zap.Any("comments", out))

	out.SortAscending()
	return out, nextPage, nil
}

// GetCommentsByImpartWealthID takes an inputImpartWealthID and optional limit and offset and returns
// a list of comments for that impartWealthId
func (d *dynamo) GetCommentsByImpartWealthID(impartWealthID string, limit int64, offset time.Time) (models.Comments, error) {
	out := make(models.Comments, 0)
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
		TableName:                 aws.String(d.getTableNameForEnvironment(commentTableName)),
		ConsistentRead:            aws.Bool(false),
		Limit:                     aws.Int64(limit),
		ScanIndexForward:          aws.Bool(false),
		IndexName:                 aws.String(gsiCommentImpartID),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	resp, err := d.Query(input)
	if err != nil {
		d.Logger.Error("error querying for Comments",
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
		d.Logger.Error("error unmarshalling dynamodb attributes",
			zap.String("impartWealthId", impartWealthID),
			zap.Error(err))
		return out, err
	}

	d.Logger.Debug("retrieved", zap.Any("comments", out))

	return out, nil
}

func commentKey(postID, commentID string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"postId": {
			S: aws.String(postID),
		},
		"commentId": {
			S: aws.String(commentID),
		},
	}
}

func (d *dynamo) IncrementDecrementComment(postID, commentID, colName string, subtract bool) error {
	return IncrementDecrement(IncrementDecrementInput{
		DynamoDBAPI: d.DynamoDBAPI,
		TableName:   d.getTableNameForEnvironment(commentTableName),
		ColumnName:  colName,
		Subtract:    subtract,
		Logger:      d.Logger,
		Key:         commentKey(postID, commentID),
	})
}

func (d *dynamo) DeleteComment(postID, commentID string) error {
	err := d.DeleteTracks(commentID)
	if err != nil && err != impart.ErrNotFound {
		d.Logger.Debug("")
		return err
	}

	_, err = d.DynamoDBAPI.DeleteItem(&dynamodb.DeleteItemInput{
		Key:       commentKey(postID, commentID),
		TableName: aws.String(d.getTableNameForEnvironment(commentTableName)),
	})

	return handleAWSErr(err)
}

func (d *dynamo) DeleteComments(postID string) error {
	var eg errgroup.Group
	comments, nextPage, err := d.GetComments(postID, 100, nil)
	if err != nil {
		return err
	}

	pageNumber := 1
	for true {
		d.Logger.Debug("Received delete request",
			zap.Int("pageNumber", pageNumber),
			zap.String("postID", postID),
			zap.Int("number of comments page", len(comments)),
			zap.Any("nextPage", nextPage))

		for i := range comments {
			commentID := comments[i].CommentID
			eg.Go(func() error {
				if deleteErr := d.DeleteComment(postID, commentID); deleteErr != nil {
					d.Logger.Error("Error deleting comment", zap.String("postId", postID), zap.String("commentId", commentID))
					return deleteErr
				}
				d.Logger.Debug("deleted comment", zap.String("postId", postID), zap.String("commentId", commentID))
				return nil
			})
		}

		if nextPage == nil {
			break
		}

		comments, nextPage, err = d.GetComments(postID, 100, nextPage)
		if err != nil && err != impart.ErrNotFound {
			return err
		}

		pageNumber++
	}

	return eg.Wait()
}

func (d *dynamo) DeleteCommentsBatch(postID string) error {
	maxDynamoDeleteSize := int64(25)

	tableName := d.getTableNameForEnvironment(commentTableName)
	var DeleteBatchFunc func(comments models.Comments) error

	DeleteBatchFunc = func(comments models.Comments) error {
		requests := make([]*dynamodb.WriteRequest, len(comments))
		for i, c := range comments {
			requests[i] = &dynamodb.WriteRequest{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: commentKey(c.PostID, c.CommentID),
				},
			}
		}
		b := &backoff.Backoff{
			Jitter: true,
			Min:    time.Millisecond * 100,
			Max:    time.Second * 5,
		}
		isFirstRun := true
		for len(requests) > 0 {

			if !isFirstRun {
				backoffTime := b.Duration()
				d.Logger.Info("not able to delete batch in 1 request - using exponential backoff", zap.Duration("duration", backoffTime), zap.Int("itemsRemaining", len(requests)))
				time.Sleep(backoffTime)
			}
			input := &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					tableName: requests,
				},
			}

			resp, err := d.BatchWriteItem(input)
			if err != nil {
				d.Logger.Error("error trying to batch delete user tracked items", zap.Error(err))
				return err
			}

			unprocessedItems, ok := resp.UnprocessedItems[tableName]
			if !ok {
				break
			}

			requests = unprocessedItems
			isFirstRun = false
		}
		return nil
	}

	var comments models.Comments
	var eg errgroup.Group
	var nextPage *models.NextPage
	var err error

	comments, nextPage, err = d.GetComments(postID, maxDynamoDeleteSize, nil)
	if err != nil {
		return handleAWSErr(err)
	}

	//Page through all the comments
	for len(comments) > 0 {
		// For each comment, create a goroutine to delete the user tracking items in the track table
		for _, c := range comments {
			commentID := c.CommentID
			eg.Go(func() error {
				err := d.DeleteTracks(commentID)
				if err == impart.ErrNotFound {
					return nil
				}
				return err
			})
		}

		// create a goroutine to delete this batch of comments
		eg.Go(func() error {
			return DeleteBatchFunc(comments)
		})

		//If there is not another page, break the for loop
		if nextPage == nil {
			break
		}

		//otherwise, get the next page and continue
		comments, nextPage, err = d.GetComments(postID, maxDynamoDeleteSize, nil)
		if err != nil {
			return handleAWSErr(err)
		}
	}

	return handleAWSErr(eg.Wait())
}
