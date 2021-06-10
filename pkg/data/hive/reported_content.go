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

func (d *mysqlHiveData) GetReviewedContents(ctx context.Context, gpi GetPostsInput) (models.Posts, models.Comments, *models.NextPage, error) {
	var postCnt int
	var cmntCnt int

	var resultComments models.Comments
	var resultPosts models.Posts

	outOffset := &models.NextPage{
		Offset:        gpi.Offset,
		OffsetPost:    gpi.OffsetPost,
		OffsetComment: gpi.OffsetComment,
	}

	if gpi.Limit <= 0 {
		gpi.Limit = defaultPostLimit
	} else if gpi.Limit > maxPostLimit {
		gpi.Limit = maxPostLimit
	}
	orderByMod := qm.OrderBy("created_at desc, post_id desc")

	queryMods := []qm.QueryMod{
		dbmodels.PostWhere.HiveID.EQ(gpi.HiveID),
		qm.Offset(gpi.OffsetPost),
		qm.Limit(gpi.Limit),
		orderByMod,
		qm.Load(dbmodels.PostRels.Tags),
		qm.Load(dbmodels.PostRels.PostReactions),
		qm.Load(dbmodels.PostRels.ImpartWealth),
	}

	queryMods = append(queryMods, qm.WhereIn("exists (select * from post_reactions rectn where rectn.post_id = `post`.`post_id` and rectn.reported = ?)", 1))
	posts, err := dbmodels.Posts(queryMods...).All(ctx, d.db)
	if err != nil {
		posts = dbmodels.PostSlice{}
	}

	queryCommnt := []qm.QueryMod{
		qm.Offset(gpi.OffsetComment),
		qm.Limit(gpi.Limit),
		orderByMod,
		qm.Load(dbmodels.CommentRels.ImpartWealth),
		qm.Load(dbmodels.CommentRels.CommentReactions),
		qm.Load(dbmodels.CommentRels.Post, dbmodels.PostWhere.HiveID.EQ(gpi.HiveID)),
	}

	queryCommnt = append(queryCommnt, qm.WhereIn("exists (select * from comment_reactions cmtrec where cmtrec.comment_id = `comment`.`comment_id` and cmtrec.reported = ?)", 1))
	comment, err := dbmodels.Comments(queryCommnt...).All(ctx, d.db)
	if err != nil {
		comment = dbmodels.CommentSlice{}
	}

	if len(posts)+len(comment) <= gpi.Limit {
		cmntCnt = len(comment)
		postCnt = len(posts)
	} else if len(posts) <= gpi.Limit {
		postCnt = len(posts)
		cmntCnt = gpi.Limit - postCnt
		if cmntCnt > len(comment) {
			cmntCnt = len(comment)
		}
	} else if len(posts) > gpi.Limit {
		postCnt = gpi.Limit
		cmntCnt = 0
	} else if len(comment) > gpi.Limit {
		cmntCnt = gpi.Limit
		postCnt = 0
	} else if len(comment) <= gpi.Limit {
		cmntCnt = len(comment)
		postCnt = gpi.Limit - cmntCnt
		if postCnt > len(posts) {
			postCnt = len(posts)
		}
	}

	if cmntCnt == len(comment) {
		resultComments = models.CommentsFromDBModelSlice(comment)
	} else if cmntCnt == 0 {
		resultComments = models.Comments{}
	} else if cmntCnt != len(comment) {
		resultComments = models.CommentsWithLimit(comment, cmntCnt)
	}

	if postCnt == len(posts) {
		resultPosts = models.PostsFromDB(posts)
	} else if postCnt == 0 {
		resultPosts = models.Posts{}
	} else if postCnt != len(posts) {
		resultPosts = models.PostsWithlimit(posts, postCnt)
	}

	if cmntCnt+postCnt < gpi.Limit {
		outOffset = nil
	} else {
		outOffset.OffsetPost += postCnt
		outOffset.OffsetComment += cmntCnt
		outOffset.Offset += (cmntCnt + postCnt)
	}

	return resultPosts, resultComments, outOffset, nil

}
