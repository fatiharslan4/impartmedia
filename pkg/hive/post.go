package hive

import (
	"context"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"strings"
	"time"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const DefaultCommentLimit = 25

func (s *service) NewPost(ctx context.Context, post models.Post) (models.Post, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)

	if len(strings.TrimSpace(post.Subject)) < 2 {
		return models.Post{}, impart.NewError(impart.ErrBadRequest, "subject is less than 2 characters")
	}

	if len(strings.TrimSpace(post.Content.Markdown)) < 10 {
		return models.Post{}, impart.NewError(impart.ErrBadRequest, "post is less than 10 characters")
	}
	shouldPin := false
	if post.IsPinnedPost {
		if ctxUser.Admin {
			shouldPin = true
		} else {
			post.IsPinnedPost = false
		}
	}
	post.ImpartWealthID = ctxUser.ImpartWealthID
	dbPost := post.ToDBModel()
	dbPost.CreatedTS = impart.CurrentUTC()
	dbPost.LastCommentTS = impart.CurrentUTC()
	tagsSlice := make(dbmodels.TagSlice, len(post.TagIDs), len(post.TagIDs))
	for i, t := range post.TagIDs {
		tagsSlice[i] = &dbmodels.Tag{TagID: uint(t)}
	}
	dbPost, err := s.postData.NewPost(ctx, dbPost, tagsSlice)
	if err != nil {
		s.logger.Error("unable to create a new post", zap.Error(err))
		return models.Post{}, impart.UnknownError
	}
	if shouldPin {
		if err := s.hiveData.PinPost(ctx, dbPost.HiveID, dbPost.PostID, true); err != nil {
			s.logger.Error("couldn't pin post", zap.Error(err))
		}

	}
	p := models.PostFromDB(dbPost)
	return p, nil
}

func (s *service) EditPost(ctx context.Context, inPost models.Post) (models.Post, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)
	existingPost, err := s.postData.GetPost(ctx, inPost.PostID)
	if err != nil {
		s.logger.Error("error fetcing post trying to edit", zap.Error(err))
		return models.Post{}, impart.UnknownError
	}
	if !ctxUser.Admin && existingPost.ImpartWealthID != ctxUser.ImpartWealthID {
		return models.Post{}, impart.NewError(impart.ErrUnauthorized, "unable to edit a post that's not yours")
	}
	tagsSlice := make(dbmodels.TagSlice, len(inPost.TagIDs), len(inPost.TagIDs))
	for i, t := range inPost.TagIDs {
		tagsSlice[i] = &dbmodels.Tag{TagID: uint(t)}
	}
	p, err := s.postData.EditPost(ctx, inPost.ToDBModel(), tagsSlice)
	if err != nil {
		return models.Post{}, impart.UnknownError
	}
	return models.PostFromDB(p), nil
}

func (s *service) GetPost(ctx context.Context, postID uint64, includeComments bool) (models.Post, impart.Error) {
	defer func(start time.Time) {
		s.logger.Debug("total post retrieve time", zap.Uint64("postId", postID), zap.Duration("elapsed", time.Since(start)))
	}(time.Now())

	var out models.Post
	var eg errgroup.Group

	var dbPost *dbmodels.Post
	eg.Go(func() error {
		defer func(start time.Time) {
			s.logger.Debug("single post retrieve time", zap.Uint64("postId", postID), zap.Duration("elapsed", time.Since(start)))
		}(time.Now())

		var err error
		dbPost, err = s.postData.GetPost(ctx, postID)
		if err != nil {
			s.logger.Error("error getting post data", zap.Error(err),
				zap.Uint64("postID", postID))
			return err
		}
		return nil
	})

	var comments dbmodels.CommentSlice
	var nextCommentPage *models.NextPage
	if includeComments {
		//var nextPage *models.NextPage
		s.logger.Debug("Received GetPost request and include comments = true",
			zap.Uint64("postID", postID), zap.Bool("comment", includeComments))

		eg.Go(func() error {
			var err error
			defer func(start time.Time) {
				s.logger.Debug("retrieved comments for post", zap.Uint64("postId", postID), zap.Duration("elapsed", time.Since(start)))
			}(time.Now())

			comments, nextCommentPage, err = s.commentData.GetComments(ctx, postID, DefaultCommentLimit, 0)
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		return out, impart.NewError(err, "error getting post")
	}

	out = models.PostFromDB(dbPost)
	out.Comments = models.CommentsFromDBModelSlice(comments)
	out.NextCommentPage = nextCommentPage

	return out, nil
}

func (s *service) GetPosts(ctx context.Context, gpi data.GetPostsInput) (models.Posts, *models.NextPage, impart.Error) {
	empty := make(models.Posts, 0, 0)
	var nextPage *models.NextPage
	var eg errgroup.Group

	var dbPosts dbmodels.PostSlice
	var pinnedPost *dbmodels.Post
	eg.Go(func() error {
		var postsError error
		dbPosts, nextPage, postsError = s.postData.GetPosts(ctx, gpi)
		if postsError == impart.ErrNotFound {
			return nil
		}
		if postsError != nil {
			s.logger.Error("unable to fetch posts", zap.Error(postsError))
		}
		return postsError
	})

	eg.Go(func() error {
		//if we're filtering on tags, or this is a secondary page request, return early.
		if len(gpi.TagIDs) > 0 || gpi.Offset > 0 {
			return nil
		}
		var pinnedError error
		hive, pinnedError := s.hiveData.GetHive(ctx, gpi.HiveID)
		if pinnedError != nil {
			s.logger.Error("unable to fetch hive", zap.Error(pinnedError))
			return pinnedError
		}
		if hive.PinnedPostID.Valid && hive.PinnedPostID.Uint64 > 0 {
			pinnedPost, pinnedError = s.postData.GetPost(ctx, hive.PinnedPostID.Uint64)
			if pinnedError != nil {
				s.logger.Error("unable to get pinned post", zap.Error(pinnedError))
			}
		}
		//returns nil so we don't fail the call if the pinned post is no longer present.
		return nil
	})

	err := eg.Wait()
	if err != nil {
		s.logger.Error("error fetching data", zap.Error(err))
		return empty, nextPage, impart.NewError(err, "error getting posts")
	}
	if dbPosts == nil {
		return empty, nil, nil
	}

	// If we have a pinned post, remove the pinned from from the returned post
	// and set the pinned post to the top of the list.
	if pinnedPost != nil {
		for i, p := range dbPosts {
			if p.PostID == pinnedPost.PostID {
				dbPosts = append(dbPosts[:i], dbPosts[i+1:]...)
			}
		}
		dbPosts = append(dbmodels.PostSlice{pinnedPost}, dbPosts...)
	}

	if len(dbPosts) == 0 {
		return models.Posts{}, nextPage, nil
	}

	out := models.PostsFromDB(dbPosts)

	return out, nextPage, nil
}

func (s *service) DeletePost(ctx context.Context, postID uint64) impart.Error {
	ctxUser := impart.GetCtxUser(ctx)
	existingPost, err := s.postData.GetPost(ctx, postID)
	if err != nil {
		s.logger.Error("error fetching post trying to edit", zap.Error(err))
		return impart.UnknownError
	}
	if !ctxUser.Admin && existingPost.ImpartWealthID != ctxUser.ImpartWealthID {
		return impart.NewError(impart.ErrUnauthorized, "unable to edit a post that's not yours")
	}

	err = s.postData.DeletePost(ctx, postID)
	if err != nil {
		return impart.UnknownError
	}

	return nil
}
