package data

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
)

var _ Hives = &mysqlHiveData{}
var _ HiveService = &mysqlHiveData{}

type mysqlHiveData struct {
	logger *zap.Logger
	db     *sql.DB
}

//counterfeiter:generate . HiveService
type HiveService interface {
	Hives
	Comments
	Posts
	UserTrack
}

func NewHiveService(db *sql.DB, logger *zap.Logger) HiveService {
	return &mysqlHiveData{
		logger: logger,
		db:     db,
	}
}

// Hives is the interface for Hive CRUD operations
type Hives interface {
	GetHives(ctx context.Context) (dbmodels.HiveSlice, error)
	GetHive(ctx context.Context, hiveID uint64) (*dbmodels.Hive, error)
	NewHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error)
	EditHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error)
	PinPost(ctx context.Context, hiveID, postID uint64, pin bool) error
	GetReportedUser(ctx context.Context, posts models.Posts) (models.Posts, error)
	GetReviewedPosts(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetUnreviewedReportedPosts(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetPostsWithUnreviewedComments(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetPostsWithReviewedComments(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error)
}

func (d *mysqlHiveData) GetHives(ctx context.Context) (dbmodels.HiveSlice, error) {
	ctxUser := impart.GetCtxUser(ctx)
	if ctxUser == nil {
		return dbmodels.HiveSlice{}, impart.UnknownError
	}
	if !ctxUser.Admin {
		return ctxUser.R.MemberHiveHives, nil
	}
	return dbmodels.Hives().All(ctx, d.db)
}

func (d *mysqlHiveData) GetHive(ctx context.Context, hiveID uint64) (*dbmodels.Hive, error) {
	return dbmodels.FindHive(ctx, d.db, hiveID)
}

func (d *mysqlHiveData) NewHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error) {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
		return nil, impart.ErrUnauthorized
	}

	// hive.HiveID = 0
	if err := hive.Insert(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}

	return hive, hive.Reload(ctx, d.db)
}

func (d *mysqlHiveData) EditHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error) {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
		return nil, impart.ErrUnauthorized
	}

	if _, err := hive.Update(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}

	return hive, hive.Reload(ctx, d.db)
}

// PinPost takes a hive and post id of a post ot pin or unpin
// if a post is being pinned, within the same transaction we need to (maybe) unpin the old post,
// mark the new post as pinned, and update the hive to point to the new post.
func (d *mysqlHiveData) PinPost(ctx context.Context, hiveID, postID uint64, pin bool) error {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
		return impart.ErrUnauthorized
	}
	fmt.Println("the pin post are here")
	toPin, err := dbmodels.FindPost(ctx, d.db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
		return err
	}
	hive, err := dbmodels.Hives(dbmodels.HiveWhere.HiveID.EQ(hiveID)).One(ctx, d.db)
	if err != nil {
		return err
	}

	if toPin.Pinned == pin && hive.PinnedPostID.Valid && hive.PinnedPostID.Uint64 == postID {
		return impart.ErrNoOp
	}
	if toPin.HiveID != hiveID {
		return impart.ErrBadRequest
	}

	//if this post has a pin, and doesn't match the request pin
	// that we're removing, do nothing.
	if hive.PinnedPostID.Uint64 == postID {
		return impart.ErrNoOp
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer impart.CommitRollbackLogger(tx, err, d.logger)

	if hive.PinnedPostID.Valid {
		if !pin && postID == hive.PinnedPostID.Uint64 || pin {
			if hive.PinnedPostID.Uint64 > 0 {
				existingPinnedPost, err := dbmodels.FindPost(ctx, d.db, hive.PinnedPostID.Uint64)
				if err != nil {
					return err
				}
				existingPinnedPost.Pinned = false
				if _, err := existingPinnedPost.Update(ctx, d.db, boil.Whitelist(dbmodels.PostColumns.Pinned)); err != nil {
					return err
				}
			}
		}
	}

	toPin.Pinned = pin
	_, err = toPin.Update(ctx, tx, boil.Whitelist(dbmodels.PostColumns.Pinned))

	if pin {
		hive.PinnedPostID.SetValid(postID)
		_, err = hive.Update(ctx, tx, boil.Whitelist(dbmodels.HiveColumns.PinnedPostID))
		return err
	} else {
		//unpin
		hive.PinnedPostID = null.Uint64{}
		_, err = hive.Update(ctx, tx, boil.Whitelist(dbmodels.HiveColumns.PinnedPostID))
		return err
	}
}
