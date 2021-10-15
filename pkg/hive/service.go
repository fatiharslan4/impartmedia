package hive

import (
	"context"
	"database/sql"

	"github.com/impartwealthapp/backend/internal/pkg/impart/config"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/media"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/impartwealthapp/backend/pkg/tags"
	"go.uber.org/zap"
)

var _ Service = &service{}

type Service interface {
	GetHive(ctx context.Context, hiveID uint64) (models.Hive, impart.Error)
	GetHives(ctx context.Context) (models.Hives, impart.Error)
	CreateHive(ctx context.Context, hive models.Hive) (models.Hive, impart.Error)
	EditHive(ctx context.Context, hive models.Hive) (models.Hive, impart.Error)
	HiveProfilePercentiles(ctx context.Context, hiveID uint64) (tags.TagComparisons, impart.Error)
	DeleteHive(ctx context.Context, hiveID uint64) impart.Error

	NewPost(ctx context.Context, post models.Post) (models.Post, impart.Error)
	EditPost(ctx context.Context, post models.Post) (models.Post, impart.Error)
	GetPost(ctx context.Context, postID uint64, includeComments bool) (models.Post, impart.Error)
	GetPosts(ctx context.Context, getPostsInput data.GetPostsInput) (models.Posts, *models.NextPage, impart.Error)
	Votes(ctx context.Context, vote VoteInput) (models.PostCommentTrack, impart.Error)
	DeletePost(ctx context.Context, postID uint64) impart.Error
	PinPost(ctx context.Context, hiveID, postID uint64, pin bool, isAdminActivity bool) impart.Error
	ReportPost(ctx context.Context, postId uint64, reason string, remove bool) (models.PostCommentTrack, impart.Error)
	ReviewPost(ctx context.Context, postId uint64, comment string, remove bool) (models.Post, impart.Error)
	AddPostVideo(ctx context.Context, postId uint64, ostVideo models.PostVideo, isAdminActivity bool, postHive map[uint64]uint64) (models.PostVideo, impart.Error)
	AddPostFiles(ctx context.Context, postFiles []models.File) ([]models.File, impart.Error)
	AddPostFilesDB(ctx context.Context, post *dbmodels.Post, file []models.File, isAdminActivity bool, postHive map[uint64]uint64) ([]models.File, impart.Error)
	ValidatePostFilesName(ctx context.Context, ctxUser *dbmodels.User, postFiles []models.File) []models.File
	NewPostForMultipleHives(ctx context.Context, post models.Post) impart.Error

	GetComments(ctx context.Context, postID uint64, limit, offset int) (models.Comments, *models.NextPage, impart.Error)
	GetComment(ctx context.Context, commentID uint64) (models.Comment, impart.Error)
	NewComment(ctx context.Context, comment models.Comment) (models.Comment, impart.Error)
	EditComment(ctx context.Context, comment models.Comment) (models.Comment, impart.Error)
	DeleteComment(ctx context.Context, commentID uint64) impart.Error
	ReportComment(ctx context.Context, commentID uint64, reason string, remove bool) (models.PostCommentTrack, impart.Error)
	ReviewComment(ctx context.Context, commentId uint64, comment string, remove bool) (models.Comment, impart.Error)

	SendCommentNotification(input models.CommentNotificationInput) impart.Error
	SendPostNotification(input models.PostNotificationInput) impart.Error

	GetReportedUser(ctx context.Context, posts models.Posts) (models.Posts, error)
	GetReportedContents(ctx context.Context, getInput data.GetReportedContentInput) (models.PostComments, *models.NextPage, error)

	UploadFile(files []models.File) error

	EditBulkPostDetails(ctx context.Context, postUpdate models.PostUpdate) *models.PostUpdate
	HiveBulkOperations(ctx context.Context, hiveUpdate models.HiveUpdate) *models.HiveUpdate

	CreateHiveRule(ctx context.Context, hiveRule models.HiveRule) (*models.HiveRule, impart.Error)
	GetHiveRules(ctx context.Context, gpi models.GetHiveInput) (models.HiveRuleLists, *models.NextPage, impart.Error)
	GetHivebyField(ctx context.Context, hiveName string) (*dbmodels.Hive, error)
}

const maxNotificationLength = 512

type service struct {
	logger              *zap.Logger
	hiveData            data.Hives
	postData            data.Posts
	commentData         data.Comments
	reactionData        data.UserTrack
	profileData         profiledata.Store
	notificationService impart.NotificationService
	db                  *sql.DB
	MediaStorage        media.StorageConfigurations
}

// New creates a new Hive Service
func New(cfg *config.Impart, db *sql.DB, logger *zap.Logger, media media.StorageConfigurations) Service {
	hd := data.NewHiveService(db, logger)
	var notificationSvc impart.NotificationService
	if cfg.Env == config.Local {
		notificationSvc = impart.NewNoopNotificationService()
	} else {
		notificationSvc = impart.NewImpartNotificationService(db, cfg.Env.String(), cfg.Region, cfg.IOSNotificationARN, logger)
	}
	profileData := profiledata.NewMySQLStore(db, logger, notificationSvc)
	svc := &service{
		logger:              logger,
		db:                  db,
		hiveData:            hd,
		postData:            hd,
		commentData:         hd,
		reactionData:        hd,
		notificationService: notificationSvc,
		profileData:         profileData,
		MediaStorage:        media,
	}

	return svc

}
