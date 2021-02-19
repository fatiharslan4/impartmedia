package hive

import (
	"strings"
	"time"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/tags"
	"go.uber.org/zap"
)

type Service interface {
	GetHive(authID, hiveID string) (models.Hive, impart.Error)
	GetHives(authID string) (models.Hives, impart.Error)
	CreateHive(authID string, hive models.Hive) (models.Hive, impart.Error)
	EditHive(authID string, hive models.Hive) (models.Hive, impart.Error)
	HiveProfilePercentiles(profileID, hiveID, authenticationId string) (tags.TagComparisons, impart.Error)

	NewPost(post models.Post, authenticationID string) (models.Post, impart.Error)
	EditPost(post models.Post, authenticationID string) (models.Post, impart.Error)
	GetPost(hiveID, postID string, consistentRead bool, authenticationID string) (models.Post, impart.Error)
	GetPosts(getPostsInput data.GetPostsInput, authenticationID string) (models.Posts, *models.NextPage, impart.Error)
	Votes(vote VoteInput, authID string) (models.PostCommentTrack, impart.Error)
	//DownVotes(hiveID, postID, commentID string, subtract bool, authenticationID string) impart.Error
	CommentCount(hiveID, postID string, subtract bool, authenticationID string) impart.Error
	DeletePost(hiveID, postID string, authenticationID string) impart.Error
	PinPost(hiveID, postID, authenticationID string, pin bool) impart.Error

	GetComments(hiveID, postID string, limit int64, nextPage *models.NextPage, authenticationID string) (models.Comments, *models.NextPage, impart.Error)
	//GetCommentsByImpartWealthID(impartWealthID string, limit int64, offset time.Time, authenticationID string) (models.Comments, impart.Error)
	GetComment(hiveID, postID, commentID string, consistentRead bool, authenticationID string) (models.Comment, impart.Error)
	NewComment(comment models.Comment, authenticationID string) (models.Comment, impart.Error)
	EditComment(comment models.Comment, authenticationID string) (models.Comment, impart.Error)
	DeleteComment(postID, commentID string, authenticationID string) impart.Error

	Logger() *zap.Logger
}

const maxNotificationLength = 512

// New creates a new Hive Service
func New(region, dynamoEndpoint, environment, platformApplicationARN string, logger *zap.Logger) Service {
	start := time.Now()
	defer func(start time.Time) {
		logger.Debug("created new complete hive service", zap.Duration("elapsed", time.Since(start)))
	}(start)

	hiveStore, err := data.NewHiveData(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}
	logger.Debug("created new hive data service", zap.Duration("elapsed", time.Since(start)))

	start = time.Now()
	postStore, err := data.NewPostData(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}
	logger.Debug("created new post data service", zap.Duration("elapsed", time.Since(start)))

	start = time.Now()
	profileStore, err := profiledata.New(region, dynamoEndpoint, environment, logger.Sugar())
	if err != nil {
		panic(err)
	}
	logger.Debug("created new profile data service", zap.Duration("elapsed", time.Since(start)))

	start = time.Now()
	commentStore, err := data.NewCommentData(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}
	logger.Debug("created new comment data service", zap.Duration("elapsed", time.Since(start)))

	start = time.Now()
	trackStore, err := data.NewContentTrack(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}
	logger.Debug("created new content track data service", zap.Duration("elapsed", time.Since(start)))

	start = time.Now()
	var notificationSvc impart.NotificationService
	if strings.Contains(dynamoEndpoint, "localhost") || strings.Contains(dynamoEndpoint, "127.0.0.1") {
		notificationSvc = impart.NewNoopNotificationService()
	} else {
		notificationSvc = impart.NewImpartNotificationService(environment, region, platformApplicationARN, logger)
	}
	logger.Debug("created new notification service", zap.Duration("elapsed", time.Since(start)))

	return &service{
		logger:              logger,
		hiveData:            hiveStore,
		postData:            postStore,
		profileData:         profileStore,
		commentData:         commentStore,
		trackStore:          trackStore,
		notificationService: notificationSvc,
	}
}

//TODO: Refactor this...maybe.
func NewWithProfile(region, dynamoEndpoint, environment, platformApplicationARN string, logger *zap.Logger, profileStore profiledata.Store) service {
	hiveStore, err := data.NewHiveData(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}

	postStore, err := data.NewPostData(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}

	commentStore, err := data.NewCommentData(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}

	trackStore, err := data.NewContentTrack(region, dynamoEndpoint, environment, logger)
	if err != nil {
		panic(err)
	}

	var notificationSvc impart.NotificationService
	if strings.Contains(dynamoEndpoint, "localhost") || strings.Contains(dynamoEndpoint, "127.0.0.1") {
		notificationSvc = impart.NewNoopNotificationService()
	} else {
		notificationSvc = impart.NewImpartNotificationService(environment, region, platformApplicationARN, logger)
	}

	return service{
		logger:              logger,
		hiveData:            hiveStore,
		postData:            postStore,
		profileData:         profileStore,
		commentData:         commentStore,
		trackStore:          trackStore,
		notificationService: notificationSvc,
	}
}

type service struct {
	logger              *zap.Logger
	hiveData            data.Hives
	postData            data.Posts
	profileData         profiledata.Store
	commentData         data.Comments
	trackStore          data.UserTrack
	notificationService impart.NotificationService
	tableEnvironment    string
}
