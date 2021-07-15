package data

import (
	"context"
	"database/sql"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var _ Comments = &mysqlHiveData{}

// Comments is the interface for Hive Comment CRUD operations
type Comments interface {
	GetComment(ctx context.Context, commentId uint64) (*dbmodels.Comment, error)
	GetComments(ctx context.Context, postID uint64, limit, offset int) (dbmodels.CommentSlice, *models.NextPage, error)
	NewComment(ctx context.Context, comment *dbmodels.Comment) (*dbmodels.Comment, error)
	EditComment(ctx context.Context, comment *dbmodels.Comment) (*dbmodels.Comment, error)
	DeleteComment(ctx context.Context, commentID uint64) error
}

// GetPost gets a single post and it's associated content from dynamodb.
func (d *mysqlHiveData) GetComment(ctx context.Context, commentID uint64) (*dbmodels.Comment, error) {
	ctxUser := impart.GetCtxUser(ctx)
	var commment dbmodels.Comment

	// c, err := dbmodels.Comments(dbmodels.CommentWhere.CommentID.EQ(commentID),
	// 	qm.Load(dbmodels.CommentRels.ImpartWealth),
	// 	qm.Load(dbmodels.CommentRels.CommentReactions, dbmodels.CommentReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
	// ).One(ctx, d.db)

	// fetch the comment if it is deleted
	err := dbmodels.NewQuery(
		qm.Select("*"),
		qm.From("comment"),
		qm.Where("comment_id = ?", commentID),
		qm.Load(dbmodels.CommentRels.ImpartWealth),
		qm.Load(dbmodels.CommentRels.CommentReactions, dbmodels.CommentReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
	).Bind(ctx, d.db, &commment)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}
	c := &commment
	return c, nil
}

// GetPosts takes a set GetPostsInput, and decides based on this input how to query DynamoDB.
func (d *mysqlHiveData) GetComments(ctx context.Context, postID uint64, limit int, offset int) (dbmodels.CommentSlice, *models.NextPage, error) {
	outOffset := &models.NextPage{
		Offset: offset,
	}
	ctxUser := impart.GetCtxUser(ctx)

	if limit <= 0 {
		limit = defaultCommentLimit
	} else if limit > maxPostLimit {
		limit = maxPostLimit
	}
	orderByMod := qm.OrderBy("comment_id desc")
	comments, err := dbmodels.Comments(dbmodels.CommentWhere.PostID.EQ(postID),
		qm.Offset(offset),
		qm.Limit(limit),
		orderByMod,
		qm.Load(dbmodels.CommentRels.ImpartWealth),
		qm.Load(dbmodels.CommentRels.CommentReactions, dbmodels.CommentReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
	).All(ctx, d.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbmodels.CommentSlice{}, nil, nil
		}
		return dbmodels.CommentSlice{}, nil, err
	}

	if len(comments) < limit {
		outOffset = nil
	} else {
		outOffset.Offset += len(comments)
	}

	return comments, outOffset, nil
}

// NewComment Creates a new Comment for a comments in DynamoDB
func (d *mysqlHiveData) NewComment(ctx context.Context, comment *dbmodels.Comment) (*dbmodels.Comment, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer impart.CommitRollbackLogger(tx, err, d.logger)
	post, err := dbmodels.Posts(dbmodels.PostWhere.PostID.EQ(comment.PostID), qm.For("UPDATE")).One(ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}

	if err := comment.Insert(ctx, tx, boil.Infer()); err != nil {
		return nil, err
	}

	post.CommentCount++
	post.LastCommentTS = comment.CreatedAt
	_, err = post.Update(ctx, tx, boil.Whitelist(dbmodels.PostColumns.CommentCount, dbmodels.PostColumns.LastCommentTS))
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return d.GetComment(ctx, comment.CommentID)
}

// EditComment takes an incoming comments, and modifies the DynamoDB record to match.
func (d *mysqlHiveData) EditComment(ctx context.Context, comment *dbmodels.Comment) (*dbmodels.Comment, error) {
	existingComment, err := dbmodels.FindComment(ctx, d.db, comment.CommentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}

	existingComment.Content = comment.Content
	if _, err := existingComment.Update(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}
	return d.GetComment(ctx, existingComment.CommentID)
}

func (d *mysqlHiveData) DeleteComment(ctx context.Context, commentID uint64) error {
	existingComment, err := d.GetComment(ctx, commentID)
	if err != nil {
		return err
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer impart.CommitRollbackLogger(tx, err, d.logger)

	_, err = existingComment.Delete(ctx, tx, false)
	if err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
	}

	post, err := dbmodels.FindPost(ctx, tx, existingComment.PostID)
	if err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
		return err
	}
	post.CommentCount--
	_, err = post.Update(ctx, tx, boil.Whitelist(dbmodels.PostColumns.CommentCount))
	if err != nil {
		return err
	}
	return tx.Commit()
}
