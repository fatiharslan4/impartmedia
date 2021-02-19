package data

import (
	"strings"
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

const userTrackTableName = "post_comment_track"
const upVotedColumnName = "upVoted"
const downVotedColumnName = "downVoted"
const votedDateTimeColumnName = "votedDatetime"
const savedColumnName = "saved"
const gsiUserTracKContentID = "gsi_contentId"

// UserTrack is the interface for tracking user's interaction with peices of content within the hive.
type UserTrack interface {
	GetUserTrack(impartWealthID, contentID string, consistentRead bool) (models.PostCommentTrack, error)
	GetUserTrackForContent(impartWealthID string, contentIDs []string) (map[string]models.PostCommentTrack, error)
	GetContentTrack(contentID string, limit int64, offsetImpartWealthID string) ([]models.PostCommentTrack, error)
	AddUpVote(impartWealthID, contentID, hiveID, postID string) error
	AddDownVote(impartWealthID, contentID, hiveID, postID string) error
	TakeUpVote(impartWealthID, contentID, hiveID, postID string) error
	TakeDownVote(impartWealthID, contentID, hiveID, postID string) error
	Save(impartWealthID, contentID, hiveID, postID string) error
	DeleteTracks(contentID string) error
}

// NewContentTrack returns an implementation of data.Comments interface
func NewContentTrack(region, endpoint, environment string, logger *zap.Logger) (UserTrack, error) {
	return newDynamo(region, endpoint, environment, logger)
}

// GetUserTrack gets a single piece of tracked content for a user
func (d *dynamo) GetUserTrack(impartWealthID, contentID string, consistentRead bool) (models.PostCommentTrack, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"impartWealthId": {
				S: aws.String(impartWealthID),
			},
			"contentId": {
				S: aws.String(contentID),
			},
		},
		TableName:      aws.String(d.getTableNameForEnvironment(userTrackTableName)),
		ConsistentRead: aws.Bool(consistentRead),
	}

	resp, err := d.GetItem(input)
	if err = handleAWSErr(err); err != nil {
		if err != impart.ErrNotFound {
			d.Logger.Error("error getting item from dynamodb", zap.Error(err))
		}
		return models.PostCommentTrack{}, err
	}

	if resp.Item == nil {
		d.Logger.Debug("get item return null", zap.Error(err))
		return models.PostCommentTrack{}, impart.ErrNotFound
	}

	var out models.PostCommentTrack
	err = dynamodbattribute.UnmarshalMap(resp.Item, &out)
	if err != nil {
		d.Logger.Error("Error trying to unmarshal attribute", zap.Error(err))
		return models.PostCommentTrack{}, err
	}
	d.Logger.Debug("retrieved", zap.Any("postCommentTrack", out))

	return out, nil
}

// GetUserTrackForContent gets a series of tracked content for a given user, based on a list of contentIDs
func (d *dynamo) GetUserTrackForContent(impartWealthID string, contentIDs []string) (map[string]models.PostCommentTrack, error) {
	out := make(map[string]models.PostCommentTrack)

	if len(contentIDs) == 0 {
		return out, nil
	}
	if len(contentIDs) > 100 {
		d.Logger.Error("Cannot get data for more than 100 pieces of content")
		return out, impart.ErrBadRequest
	}

	tbl := d.getTableNameForEnvironment(userTrackTableName)
	var keys []map[string]*dynamodb.AttributeValue
	for _, cid := range contentIDs {
		keys = append(keys, userTrackKey(impartWealthID, cid))
	}
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			tbl: {
				Keys: keys,
			},
		},
	}

	resp, err := d.BatchGetItem(input)
	if err = handleAWSErr(err); err != nil {
		if err != impart.ErrNotFound {
			d.Logger.Error("error querying track table",
				zap.String("impartWealthID", impartWealthID),
				zap.Error(err))
		}
		return out, err
	}

	values, ok := resp.Responses[tbl]
	if !ok {
		d.Logger.Debug("get items returned nil",
			zap.String("impartWealthID", impartWealthID),
			zap.Error(err))
		return out, impart.ErrNotFound
	}

	trackList := make([]models.PostCommentTrack, 0)

	err = dynamodbattribute.UnmarshalListOfMaps(values, &trackList)
	if err != nil {
		d.Logger.Error("Error trying to unmarshal userTrackItem",
			zap.String("impartWealthID", impartWealthID),
			zap.Error(err))
		return out, err
	}

	for _, t := range trackList {
		out[t.ContentID] = t
	}

	d.Logger.Debug("retrieved",
		zap.String("impartWealthID", impartWealthID),
		zap.Any("postCommentTrack", out))

	return out, nil
}

// GetContentTrack gets all of the users which interacted with the input content ID
func (d *dynamo) GetContentTrack(contentID string, limit int64, offsetImpartWealthID string) ([]models.PostCommentTrack, error) {
	out := make([]models.PostCommentTrack, 0)
	var err error

	kc := expression.Key("contentId").Equal(expression.Value(contentID))

	expr, err := expression.NewBuilder().WithKeyCondition(kc).Build()
	if err != nil {
		d.Logger.Error("Error building DynamoDB Condition", zap.Error(err))
		return out, err
	}

	if limit <= 0 {
		limit = DefaultLimit
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(d.getTableNameForEnvironment(userTrackTableName)),
		ConsistentRead:            aws.Bool(false),
		Limit:                     aws.Int64(limit),
		ScanIndexForward:          aws.Bool(true),
		IndexName:                 aws.String(gsiUserTracKContentID),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	if strings.TrimSpace(offsetImpartWealthID) != "" {
		input.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"contentId": {
				S: aws.String(contentID),
			},
			"impartWealthId": {
				S: aws.String(offsetImpartWealthID),
			},
		}
	}

	resp, err := d.Query(input)
	if err = handleAWSErr(err); err != nil {
		if err != impart.ErrNotFound {
			d.Logger.Error("error querying for content tracking",
				zap.String("contentID", contentID),
				zap.Error(err))
		}
		return out, err
	}

	if resp.Items == nil {
		d.Logger.Error("get items returned nil",
			zap.String("contentID", contentID),
			zap.Error(err))
		return out, impart.ErrNotFound
	}

	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &out)
	if err != nil {
		d.Logger.Error("error unmarshalling dynamodb attributs",
			zap.String("contentID", contentID),
			zap.Error(err))
		return out, err
	}

	d.Logger.Debug("retrieved", zap.Any("contentTracks", out))

	return out, nil
}

// AddUpVote sets the user tracked "upvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *dynamo) AddUpVote(impartWealthID, contentID, hiveID, postID string) error {
	return d.vote(true, true, impartWealthID, contentID, hiveID, postID)
}

// AddDownVote sets the user tracked "downvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *dynamo) AddDownVote(impartWealthID, contentID, hiveID, postID string) error {
	return d.vote(false, true, impartWealthID, contentID, hiveID, postID)
}

// TakeUpVote clears the user tracked "upvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *dynamo) TakeUpVote(impartWealthID, contentID, hiveID, postID string) error {
	return d.vote(true, false, impartWealthID, contentID, hiveID, postID)
}

// TakeDownVote clears the user tracked "downvote" of the input content ID; throws impart.ErrNoOp when no action is taken.
func (d *dynamo) TakeDownVote(impartWealthID, contentID, hiveID, postID string) error {
	return d.vote(false, false, impartWealthID, contentID, hiveID, postID)
}

// vote tracks the vote status for a user on a given piece of content, and updates the vote counts on the
// content accordingly.
// if contentID == postID or postID is blank, then it is assumed that contentID is a postID.
// if postID is not empty, and contentID != postID, then it is assumed a contentID is a commentID.
func (d *dynamo) vote(upVote, increment bool, impartWealthID, contentID, hiveID, postID string) (err error) {
	var upVoteChange int
	var downVoteChange int
	var commentID string
	postID = strings.TrimSpace(postID)

	if postID != "" && contentID != postID {
		commentID = contentID
	}

	//Mutually exclusive
	downVote := !upVote
	decrement := !increment

	trackExists := true
	t, err := d.GetUserTrack(impartWealthID, contentID, true)
	if err != nil {
		if err != impart.ErrNotFound {
			return err
		}
		trackExists = false
	}

	if !trackExists {
		d.Logger.Debug("Tracking for this content does not exist, creating")
		err = d.newUserTrack(models.PostCommentTrack{
			ImpartWealthID: impartWealthID,
			ContentID:      contentID,
			HiveID:         hiveID,
			PostID:         postID,
			VotedDatetime:  impart.CurrentUTC(),
			//if it's an upvote and we're incrementing, set upvoted
			UpVoted: upVote && increment,
			//if it's a downvote and we're decrementing, set downvoted.
			DownVoted: downVote && increment,
		})
		if err != nil {
			return err
		}

		if upVote == true {
			upVoteChange = 1
		} else {
			downVoteChange = 1
		}

	} else {

		if upVoteChange, downVoteChange = voteDelta(t, upVote, increment); upVoteChange == 0 && downVoteChange == 0 {
			d.Logger.Debug("Track noop")
			return impart.ErrNoOp
		}

		var upd expression.UpdateBuilder

		// if clearing vote
		if decrement {
			//clear the upvote
			if upVote {
				upd = expression.Set(expression.Name(upVotedColumnName), expression.Value(false)).
					Set(expression.Name(votedDateTimeColumnName), expression.Value(time.Time{}))
				//clear the downvote
			} else {
				upd = expression.Set(expression.Name(downVotedColumnName), expression.Value(false)).
					Set(expression.Name(votedDateTimeColumnName), expression.Value(time.Time{}))
			}
			//we're doing an initial vote with existing track object, or swapping votes
		} else {
			upd = expression.Set(expression.Name(upVotedColumnName), expression.Value(upVote)).
				Set(expression.Name(downVotedColumnName), expression.Value(downVote)).
				Set(expression.Name(votedDateTimeColumnName), expression.Value(impart.CurrentUTC()))
		}

		input := UpdateItemPropertyInput{
			DynamoDBAPI: d,
			TableName:   d.getTableNameForEnvironment(userTrackTableName),
			Update:      upd,
			Logger:      d.Logger,
			Condition:   nil,
			Key:         userTrackKey(impartWealthID, contentID),
		}
		d.Logger.Debug("updating existing tracked content", zap.Any("existing", t), zap.Any("update", input.Update), zap.Any("key", input.Key))

		err = UpdateItemProperty(input)
		if err != nil {
			return err
		}
	}

	return d.trackCounts(hiveID, postID, commentID, upVoteChange, downVoteChange)

}

func (d *dynamo) Save(impartWealthID, contentID, hiveID, postID string) error {
	trackExists := true
	_, err := d.GetUserTrack(impartWealthID, contentID, true)
	if err != nil {
		if err != impart.ErrNotFound {
			return err
		}
		trackExists = false
	}

	if !trackExists {
		return d.newUserTrack(models.PostCommentTrack{
			ImpartWealthID: impartWealthID,
			ContentID:      contentID,
			HiveID:         hiveID,
			PostID:         postID,
			Saved:          true,
		})
	}

	input := UpdateItemPropertyInput{
		DynamoDBAPI: d,
		TableName:   d.getTableNameForEnvironment(userTrackTableName),
		Update:      expression.Set(expression.Name(savedColumnName), expression.Value(true)),
		Logger:      d.Logger,
		Condition:   nil,
		Key:         userTrackKey(impartWealthID, contentID),
	}

	return UpdateItemProperty(input)
}

func (d *dynamo) newUserTrack(track models.PostCommentTrack) error {
	d.Logger.Debug("Received new user Track request", zap.Any("trackRequest", track))
	var err error

	item, err := dynamodbattribute.MarshalMap(track)
	if err != nil {
		return err
	}

	cb := expression.Name("impartWealthId").AttributeNotExists().
		And(expression.Name("contentId").AttributeNotExists())

	expr, err := expression.NewBuilder().
		WithCondition(cb).
		Build()
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:                     item,
		ReturnConsumedCapacity:   aws.String("NONE"),
		TableName:                aws.String(d.getTableNameForEnvironment(userTrackTableName)),
		ConditionExpression:      expr.Condition(),
		ExpressionAttributeNames: expr.Names(),
		ReturnValues:             aws.String("NONE"),
	}

	_, err = d.PutItem(input)
	return err
}

func userTrackKey(impartWealthID, contentID string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"impartWealthId": {
			S: aws.String(impartWealthID),
		},
		"contentId": {
			S: aws.String(contentID),
		},
	}
}

// voteDelta checks whether the track is a noop or not.  a false upvote is a downvote, and false increment is a decrement.
func voteDelta(track models.PostCommentTrack, isUpvote, increment bool) (upVoteDelta, downVoteDelta int) {

	if isUpvote {
		if !track.UpVoted && !track.DownVoted && increment {
			return 1, 0
		}
		if !track.UpVoted && track.DownVoted && increment {
			return 1, -1
		}
		if track.UpVoted && !increment {
			return -1, 0
		}
	} else {
		//is a downvote
		if !track.DownVoted && !track.UpVoted && increment {
			return 0, 1
		}

		if !track.DownVoted && track.UpVoted && increment {
			return -1, 1
		}

		if track.DownVoted && !increment {
			return 0, -1
		}
	}

	return 0, 0

}

func (d *dynamo) DeleteTracks(contentID string) error {
	maxDynamoDeleteSize := int64(25)

	tableName := d.getTableNameForEnvironment(userTrackTableName)
	var DeleteBatchFunc func(trackedItems []models.PostCommentTrack) error

	DeleteBatchFunc = func(trackedItems []models.PostCommentTrack) error {
		requests := make([]*dynamodb.WriteRequest, len(trackedItems))
		for i, t := range trackedItems {
			requests[i] = &dynamodb.WriteRequest{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: userTrackKey(t.ImpartWealthID, t.ContentID),
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

	var tracks []models.PostCommentTrack
	var eg errgroup.Group
	var err error
	offsetImpartWealthID := ""

	tracks, err = d.GetContentTrack(contentID, maxDynamoDeleteSize, offsetImpartWealthID)
	if err != nil {
		return handleAWSErr(err)
	}

	for len(tracks) > 0 {

		eg.Go(func() error {
			return DeleteBatchFunc(tracks)
		})

		for _, t := range tracks {
			if t.ImpartWealthID > offsetImpartWealthID {
				offsetImpartWealthID = t.ImpartWealthID
			}
		}

		tracks, err = d.GetContentTrack(contentID, maxDynamoDeleteSize, offsetImpartWealthID)
		if err != nil {
			return handleAWSErr(err)
		}
	}

	return handleAWSErr(eg.Wait())
}

func (d *dynamo) trackCounts(hiveID, postID, commentID string, upVoteChange, downVoteChange int) error {

	var eg errgroup.Group

	if upVoteChange != 0 {
		eg.Go(func() error {
			subtract := upVoteChange < 0
			var egErr error
			if commentID != "" {
				egErr = d.IncrementDecrementComment(postID, commentID, UpVoteCountColumnName, subtract)
			} else {
				egErr = d.IncrementDecrementPost(hiveID, postID, UpVoteCountColumnName, subtract)
			}

			if egErr != nil {
				d.Logger.Error("error tracking upvotes for post or comments", zap.Error(egErr))
			}
			return egErr
		})
	}

	if downVoteChange != 0 {
		eg.Go(func() error {
			var egErr error
			subtract := downVoteChange < 0
			if commentID != "" {
				egErr = d.IncrementDecrementComment(postID, commentID, DownVoteCountColumnName, subtract)
			} else {
				egErr = d.IncrementDecrementPost(hiveID, postID, DownVoteCountColumnName, subtract)
			}
			if egErr != nil {
				d.Logger.Error("error tracking downvotes for post or comments")
			}
			return egErr
		})
	}

	return eg.Wait()
}
