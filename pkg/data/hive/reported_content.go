package data

import (
	"context"
	"database/sql"

	// "fmt"
	"time"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// type reportedUser struct {
// 	reportedUsers []string
// }

func (d *mysqlHiveData) GetReviewedPosts(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error) {
	ctxUser := impart.GetCtxUser(ctx)
	var nextPage models.NextPage
	posts, err := dbmodels.Posts(
		dbmodels.PostWhere.HiveID.EQ(hiveId),
		dbmodels.PostWhere.ReviewedAt.GTE(null.TimeFrom(reviewDate)),
		qm.OrderBy(dbmodels.PostColumns.ReviewedAt),
		qm.Limit(100),
		qm.Offset(offset),
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.ImpartWealth),
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
	).All(ctx, d.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbmodels.PostSlice{}, nextPage, impart.ErrNotFound
		}
		return dbmodels.PostSlice{}, nextPage, err
	}
	nextPage.Offset = offset + len(posts)
	return posts, nextPage, nil
}

func (d *mysqlHiveData) GetReportedUser(ctx context.Context, posts models.Posts) (models.Posts, error) {
	var postIds []uint64
	var indexes map[uint64]int
	indexes = make(map[uint64]int)
	for index, value := range posts {
		if value.ReportedCount > 0 {
			postIds = append(postIds, (value.PostID))
			indexes[value.PostID] = index
		}
	}
	PostReactions, err := dbmodels.PostReactions(qm.Select("impart_wealth_id", "post_id"), dbmodels.PostReactionWhere.PostID.IN(postIds)).All(ctx, d.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return posts, nil
		}
		return models.Posts{}, err
	}
	for _, value := range PostReactions {
		if postID, ok := indexes[value.PostID]; ok {
			posts[postID].ReportedUsers = append(posts[postID].ReportedUsers, models.ReportedUser{
				ImpartWealthID: value.ImpartWealthID})

		}
	}
	return posts, nil
}

func (d *mysqlHiveData) GetUnreviewedReportedPosts(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error) {
	ctxUser := impart.GetCtxUser(ctx)
	var nextPage models.NextPage
	posts, err := dbmodels.Posts(
		dbmodels.PostWhere.HiveID.EQ(hiveId),
		dbmodels.PostWhere.ReviewedAt.IsNull(),
		dbmodels.PostWhere.ReportedCount.GT(0),
		qm.OrderBy(dbmodels.PostColumns.LastCommentTS),
		qm.Limit(100),
		qm.Offset(offset),
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.ImpartWealth),
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
	).All(ctx, d.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbmodels.PostSlice{}, nextPage, impart.ErrNotFound
		}
		return dbmodels.PostSlice{}, nextPage, err
	}
	nextPage.Offset = offset + len(posts)
	return posts, nextPage, nil
}

func (d *mysqlHiveData) GetPostsWithUnreviewedComments(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error) {
	var empty dbmodels.PostSlice
	var nextPage models.NextPage
	ctxUser := impart.GetCtxUser(ctx)
	type postComment struct {
		Post    dbmodels.Post    `boil:",bind"`
		Comment dbmodels.Comment `boil:",bind"`
	}

	var postComments []*postComment

	err := queries.Raw(`
SELECT distinct
    post_id, comment_id
FROM (
     select p.post_id,
            c.comment_id
     from comment c
     join post p
        on p.post_id = c.post_id
     where p.hive_id = ?
       and c.reviewed_at is not null
       and c.reported_count > 0
     order by p.last_comment_ts
     LIMIT ? OFFSET >
) as aud

	`, hiveId, 100, offset, hiveId).Bind(ctx, d.db, &postComments)

	if err != nil {
		if err == sql.ErrNoRows {
			return empty, nextPage, impart.ErrNotFound
		}
		return empty, nextPage, err
	}
	nextPage.Offset = offset + len(postComments)
	uniquePostIds := make(map[uint64]struct{})
	commentIds := make([]uint64, len(postComments), len(postComments))
	for i, pc := range postComments {
		uniquePostIds[pc.Post.PostID] = struct{}{}
		commentIds[i] = pc.Comment.CommentID
	}

	postIds := make([]uint64, len(uniquePostIds), len(uniquePostIds))
	var i int
	for k, _ := range uniquePostIds {
		postIds[i] = k
		i++
	}

	posts, err := dbmodels.Posts(
		dbmodels.PostWhere.HiveID.EQ(hiveId),
		dbmodels.PostWhere.PostID.IN(postIds),
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.ImpartWealth),
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
		qm.Load(dbmodels.PostRels.Comments, dbmodels.CommentWhere.CommentID.IN(commentIds)),
	).All(ctx, d.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbmodels.PostSlice{}, nextPage, impart.ErrNotFound
		}
		return dbmodels.PostSlice{}, nextPage, err
	}
	return posts, nextPage, nil
}

func (d *mysqlHiveData) GetPostsWithReviewedComments(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error) {
	var empty dbmodels.PostSlice
	var nextPage models.NextPage
	ctxUser := impart.GetCtxUser(ctx)
	type postComment struct {
		Post    dbmodels.Post    `boil:",bind"`
		Comment dbmodels.Comment `boil:",bind"`
	}

	var postComments []*postComment

	err := queries.Raw(`
SELECT distinct
    post_id, comment_id
FROM (
     select p.post_id,
            c.comment_id
     from comment c
     join post p
        on p.post_id = c.post_id
     where p.hive_id = ?
       and c.reviewed_at >= ?
       and c.reported_count > 0
     order by p.last_comment_ts
     LIMIT ? OFFSET >
) as aud

	`, hiveId, reviewDate, 100, offset, hiveId).Bind(ctx, d.db, &postComments)

	if err != nil {
		if err == sql.ErrNoRows {
			return empty, nextPage, impart.ErrNotFound
		}
		return empty, nextPage, err
	}
	nextPage.Offset = offset + len(postComments)
	uniquePostIds := make(map[uint64]struct{})
	commentIds := make([]uint64, len(postComments), len(postComments))
	for i, pc := range postComments {
		uniquePostIds[pc.Post.PostID] = struct{}{}
		commentIds[i] = pc.Comment.CommentID
	}

	postIds := make([]uint64, len(uniquePostIds), len(uniquePostIds))
	var i int
	for k, _ := range uniquePostIds {
		postIds[i] = k
		i++
	}

	posts, err := dbmodels.Posts(
		dbmodels.PostWhere.HiveID.EQ(hiveId),
		dbmodels.PostWhere.PostID.IN(postIds),
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.ImpartWealth),
		qm.Load(dbmodels.PostRels.PostReactions, dbmodels.PostReactionWhere.ImpartWealthID.EQ(ctxUser.ImpartWealthID)),
		qm.Load(dbmodels.PostRels.Comments, dbmodels.CommentWhere.CommentID.IN(commentIds)),
	).All(ctx, d.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbmodels.PostSlice{}, nextPage, impart.ErrNotFound
		}
		return dbmodels.PostSlice{}, nextPage, err
	}
	return posts, nextPage, nil
}
