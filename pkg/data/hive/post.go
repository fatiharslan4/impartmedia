package data

import (
	"context"
	"database/sql"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

var _ Posts = &mysqlHiveData{}

const defaultPostLimit = 100
const defaultCommentLimit = 100
const maxPostLimit = 256
const maxCommentLimit = 256

// Posts is the interface for Hive Post CRUD operations
type Posts interface {
	GetPosts(ctx context.Context, getPostsInput GetPostsInput) (dbmodels.PostSlice, *models.NextPage, error)
	//GetPostsImpartWealthID(ctx context.Context, impartWealthID string, limit int64, offset time.Time) (models.Posts, error)
	GetPost(ctx context.Context, postID uint64) (*dbmodels.Post, error)
	NewPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice) (*dbmodels.Post, error)
	EditPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice, shouldPin bool) (*dbmodels.Post, error)
	DeletePost(ctx context.Context, postID uint64) error
	GetReportedUser(ctx context.Context, posts models.Posts) (models.Posts, error)
	NewPostVideo(ctx context.Context, post *dbmodels.PostVideo) (*dbmodels.PostVideo, error)
}

// GetPost gets a single post and it's associated content
func (d *mysqlHiveData) GetPost(ctx context.Context, postID uint64) (*dbmodels.Post, error) {
	ctxUser := impart.GetCtxUser(ctx)
	// p, err := dbmodels.Posts(dbmodels.PostWhere.PostID.EQ(postID),
	// 	qm.Load(dbmodels.PostRels.Tags),
	// 	qm.Load(dbmodels.PostRels.ImpartWealth),
	// 	qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
	// ).One(ctx, d.db)

	var post dbmodels.Post
	err := dbmodels.NewQuery(
		qm.Select("*"),
		qm.From("post"),
		qm.Where("post_id = ?", postID),
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.ImpartWealth),
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
		qm.Load(dbmodels.PostRels.PostVideos),
		qm.Load(dbmodels.PostRels.PostFiles),
		qm.Load("PostFiles.FidFile"), // get files
	).Bind(ctx, d.db, &post)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}
	p := &post
	return p, nil
}

// NewPost Creates a new Post in DynamoDB
func (d *mysqlHiveData) NewPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice) (*dbmodels.Post, error) {
	err := post.Insert(ctx, d.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	err = post.SetTags(ctx, d.db, false, tags...)
	if err != nil {
		return nil, err
	}
	return d.GetPost(ctx, post.PostID)
}

// EditPost takes an incoming Post, and modifies the record to match.
func (d *mysqlHiveData) EditPost(ctx context.Context, post *dbmodels.Post, tags dbmodels.TagSlice, shouldPin bool) (*dbmodels.Post, error) {
	//you can only edit content and subject
	existing, err := dbmodels.FindPost(ctx, d.db, post.PostID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}

	if post.Content != "" && post.Content != existing.Content {
		existing.Content = post.Content
	}

	if post.Subject != "" && post.Subject != existing.Subject {
		existing.Subject = post.Subject
	}
	_, err = existing.Update(ctx, d.db, boil.Infer())

	if shouldPin && post.Pinned != existing.Pinned {
		err = d.PinPost(ctx, post.HiveID, post.PostID, post.Pinned)
	}
	if err := existing.SetTags(ctx, d.db, false, tags...); err != nil {
		return nil, err
	}
	return d.GetPost(ctx, post.PostID)

}

// GetPostsInput is the input necessary
type GetPostsInput struct {
	// HiveID is the ID that should be queried for posts
	HiveID uint64
	// Limit is the maximum number of records that should be returns.  The API can optionally return
	// less than Limit, if DynamoDB decides the items read were too large.
	Limit  int
	Offset int
	// IsLastCommentSorted Changes the sort from default of PostDatetime to LastCommentDatetime
	// Default: false
	IsLastCommentSorted bool
	// Tags is the optional list of tags to filter on
	TagIDs []int

	OffsetPost    int
	OffsetComment int
}

// GetPosts takes a set GetPostsInput, and decides based on this input how to query DynamoDB.
func (d *mysqlHiveData) GetPosts(ctx context.Context, gpi GetPostsInput) (dbmodels.PostSlice, *models.NextPage, error) {
	ctxUser := impart.GetCtxUser(ctx)
	var empty dbmodels.PostSlice
	outOffset := &models.NextPage{
		Offset: gpi.Offset,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultPostLimit
	} else if gpi.Limit > maxPostLimit {
		gpi.Limit = maxPostLimit
	}

	orderByMod := qm.OrderBy("created_at desc, post_id desc")
	if gpi.IsLastCommentSorted {
		orderByMod = qm.OrderBy("last_comment_ts desc, post_id desc")
	}
	queryMods := []qm.QueryMod{
		dbmodels.PostWhere.HiveID.EQ(gpi.HiveID),
		qm.Offset(gpi.Offset),
		qm.Limit(gpi.Limit),
		orderByMod,
		qm.Load(dbmodels.PostRels.Tags), // all the tags associated with this post
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)), // the callers reaction
		qm.Load(dbmodels.PostRels.ImpartWealth), // the user who posted
		qm.Load(dbmodels.PostRels.PostFiles),    // get post files
		qm.Load("PostFiles.FidFile"),            // get files
	}

	if len(gpi.TagIDs) > 0 {
		inParamValues := make([]interface{}, len(gpi.TagIDs), len(gpi.TagIDs))
		for i, id := range gpi.TagIDs {
			inParamValues[i] = id
		}
		queryMods = append(queryMods, qm.WhereIn("exists (select * from post_tag pt where pt.post_id = `post`.`post_id` and pt.tag_id in ?)", inParamValues...))
	}

	posts, err := dbmodels.Posts(queryMods...).All(ctx, d.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return empty, nil, nil
		}
		d.logger.Error("couldn't fetch posts from db", zap.Error(err))
		return empty, nil, err
	}
	boil.DebugMode = false
	if len(posts) < gpi.Limit {
		outOffset = nil
	} else {
		outOffset.Offset += len(posts)
	}

	return posts, outOffset, nil
}

func (d *mysqlHiveData) DeletePost(ctx context.Context, postID uint64) error {
	p, err := dbmodels.FindPost(ctx, d.db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if _, err = p.Delete(ctx, d.db, false); err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
		return err
	}
	if p.Pinned {
		q := `
UPDATE hive
	SET pinned_post_id = null
WHERE pinned_post_id = ?;`
		_, err := queries.Raw(q, postID).ExecContext(ctx, d.db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *mysqlHiveData) NewPostVideo(ctx context.Context, postVideo *dbmodels.PostVideo) (*dbmodels.PostVideo, error) {
	if err := postVideo.Insert(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}
	return postVideo, nil
}
