package hive

import (
	"reflect"
	"strings"
	"time"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (s *service) NewPost(post models.Post, authenticationID string) (models.Post, impart.Error) {
	profile, impartErr := s.validateHiveAccess(post.HiveID, authenticationID)
	if impartErr != nil {
		s.logger.Error("cannot validate the user is able to post to this hive", zap.Any("impartError", impartErr))
		return models.Post{}, impartErr
	}

	if len(strings.TrimSpace(post.Subject)) < 2 {
		return models.Post{}, impart.NewError(impart.ErrBadRequest, "subject is less than 2 characters")
	}

	if len(strings.TrimSpace(post.Content.Markdown)) < 10 {
		return models.Post{}, impart.NewError(impart.ErrBadRequest, "post is less than 10 characters")
	}

	p := models.Post{
		HiveID:              post.HiveID,
		PostID:              ksuid.New().String(),
		PostDatetime:        impart.CurrentUTC(),
		LastCommentDatetime: impart.CurrentUTC(),
		ImpartWealthID:      profile.ImpartWealthID,
		ScreenName:          profile.ScreenName,
		Subject:             post.Subject,
		Content:             post.Content,
		TagIDs:              post.TagIDs,
	}
	if profile.Attributes.Admin && post.IsPinnedPost {
		p.IsPinnedPost = true
	}

	p, err := s.postData.NewPost(p)
	if err != nil {
		s.logger.Error("error creating new post", zap.Error(err))
		return models.Post{}, impart.NewError(err, "unable to create new post")
	}

	if p.IsPinnedPost {
		impartErr = s.PinPost(p.HiveID, p.PostID, profile.AuthenticationID, true)
	}
	return p, impartErr
}

func (s *service) EditPost(inPost models.Post, authenticationID string) (models.Post, impart.Error) {
	existingPost, err := s.postData.GetPost(inPost.HiveID, inPost.PostID, false)
	if err != nil {
		return models.Post{}, impart.NewError(err, "unable to locate existing post to edit")
	}

	profile, impartErr := s.selfOrAdmin(existingPost.HiveID, existingPost.ImpartWealthID, authenticationID)
	if impartErr != nil {
		s.logger.Error("user is not authorized to edit this post",
			zap.Any("post", existingPost), zap.String("authenticationId", authenticationID))
		return models.Post{}, impartErr
	}

	if existingPost.Content.Markdown == inPost.Content.Markdown &&
		existingPost.Subject == inPost.Subject &&
		reflect.DeepEqual(existingPost.TagIDs, inPost.TagIDs) {
		return models.Post{}, impart.NewError(impart.ErrBadRequest, "content has not changed")
	}

	newEdit := models.Edit{
		Datetime:       impart.CurrentUTC(),
		ImpartWealthID: profile.ImpartWealthID,
		ScreenName:     profile.ScreenName,
	}

	existingPost.Edits = append(existingPost.Edits, newEdit)
	existingPost.Content = inPost.Content
	existingPost.TagIDs = inPost.TagIDs
	existingPost.Subject = inPost.Subject

	if profile.Attributes.Admin && inPost.IsPinnedPost && !existingPost.IsPinnedPost {
		existingPost.IsPinnedPost = true
	}

	s.logger.Debug("Editing post", zap.Any("edit", newEdit), zap.Any("updated post", existingPost))
	out, err := s.postData.EditPost(existingPost)
	if err != nil {
		return models.Post{}, impart.NewError(err, "error editing post")
	}

	if existingPost.IsPinnedPost {
		impartErr = s.PinPost(existingPost.HiveID, existingPost.PostID, profile.AuthenticationID, true)
	}

	return out, impartErr
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (s *service) GetPost(hiveID, postID string, includeComments bool, authenticationID string) (models.Post, impart.Error) {
	defer func(start time.Time) {
		s.logger.Debug("total post retrieve time", zap.String("postId", postID), zap.Duration("elapsed", time.Since(start)))
	}(time.Now())

	var out models.Post

	p, impartErr := s.validateHiveAccess(hiveID, authenticationID)
	if impartErr != nil {
		return out, impartErr
	}

	var eg errgroup.Group

	var tmpPost models.Post
	eg.Go(func() error {
		defer func(start time.Time) {
			s.logger.Debug("single post retrieve time", zap.String("postId", postID), zap.Duration("elapsed", time.Since(start)))
		}(time.Now())

		var err error
		tmpPost, err = s.postData.GetPost(hiveID, postID, false)
		if err != nil {
			s.logger.Error("error getting post data", zap.Error(err),
				zap.String("hiveID", hiveID), zap.String("postID", postID))
			return err
		}
		return nil
	})

	var tmpTracks models.PostCommentTrack
	eg.Go(func() error {
		defer func(start time.Time) {
			s.logger.Debug("post comment track retrieve time", zap.String("postId", postID), zap.Duration("elapsed", time.Since(start)))
		}(time.Now())

		var err error
		tmpTracks, err = s.trackStore.GetUserTrack(p.ImpartWealthID, postID, false)
		if err != nil && err != impart.ErrNotFound {
			s.logger.Error("error getting user track", zap.Error(err),
				zap.String("hiveID", hiveID), zap.String("postID", postID))
			return err
		}
		return nil
	})

	var tmpComments models.Comments
	if includeComments {
		//var nextPage *models.NextPage
		s.logger.Debug("Received GetPost request and include comments = true",
			zap.String("hiveID", hiveID), zap.String("postID", postID), zap.Bool("comment", includeComments))

		eg.Go(func() error {
			defer func(start time.Time) {
				s.logger.Debug("retrieved comments for post", zap.String("postId", postID), zap.Duration("elapsed", time.Since(start)))
			}(time.Now())

			tmpComments, _, impartErr = s.GetComments(hiveID, postID, data.DefaultLimit, nil, authenticationID)
			if impartErr != nil {
				return impartErr.Err()
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return out, impart.NewError(err, "error getting postst")
	}

	out = tmpPost
	out.PostCommentTrack = tmpTracks
	out.Comments = tmpComments

	return out, nil
}

func (s *service) GetPosts(gpi data.GetPostsInput, authenticationID string) (models.Posts, *models.NextPage, impart.Error) {
	var out models.Posts
	var nextPage *models.NextPage
	var pinnedPost models.Post

	p, impartErr := s.validateHiveAccess(gpi.HiveID, authenticationID)
	if impartErr != nil {
		return out, nextPage, impartErr
	}

	var eg errgroup.Group

	eg.Go(func() error {
		var postsError error
		out, nextPage, postsError = s.postData.GetPosts(gpi)
		return postsError
	})

	eg.Go(func() error {
		//if we're filtering on tags, or this is a secondary page request, return early.
		if len(gpi.TagIDs) > 0 || gpi.NextPage != nil {
			return nil
		}
		var pinnedError error
		hive, pinnedError := s.hiveData.GetHive(gpi.HiveID, false)
		if pinnedError != nil {
			return pinnedError
		}
		if hive.PinnedPostID != "" {
			pinnedPost, pinnedError = s.postData.GetPost(gpi.HiveID, hive.PinnedPostID, false)
			if pinnedError != nil {
				s.logger.Error("unable to get pinned post", zap.Error(pinnedError))
			}
		}
		//returns nil so we don't fail the call if the pinned post is no longer present.
		return nil

	})

	err := eg.Wait()
	if err != nil {
		return out, nextPage, impart.NewError(err, "error getting posts")
	}

	// If we have a pinned post, remove the pinned from from the returned post
	// and set the pinned post to the top of the list.
	if pinnedPost.PostID != "" {
		for i, p := range out {
			if p.PostID == pinnedPost.PostID {
				out = append(out[:i], out[i+1:]...)
			}
		}
		out = append(models.Posts{pinnedPost}, out...)
	}

	postIDs := make([]string, len(out))
	for i, p := range out {
		postIDs[i] = p.PostID
	}

	if len(postIDs) == 0 {
		return models.Posts{}, nextPage, nil
	}
	tracks, err := s.trackStore.GetUserTrackForContent(p.ImpartWealthID, postIDs)
	if err != nil {
		if err != impart.ErrNotFound {
			s.logger.Debug("error getting user track", zap.Error(err))
			return out, nextPage, impart.NewError(err, "error getting tracked items")
		}
	}
	for i, p := range out {
		out[i].PostCommentTrack = tracks[p.PostID]
	}

	return out, nextPage, nil
}

func (s *service) DeletePost(hiveID, postID, authenticationID string) impart.Error {
	existingPost, err := s.postData.GetPost(hiveID, postID, false)
	if err != nil {
		s.logger.Error("Error getting existing post data",
			zap.String("hiveID", hiveID),
			zap.String("postID", postID),
			zap.String("authenticationID", authenticationID))
		return impart.NewError(err, "unable to find existing post")
	}
	_, impartErr := s.selfOrAdmin(existingPost.HiveID, existingPost.ImpartWealthID, authenticationID)
	if impartErr != nil {
		s.logger.Error("user is not authorized to delete this post",
			zap.Any("post", existingPost), zap.String("authenticationId", authenticationID))
		return impartErr
	}

	err = s.postData.DeletePost(hiveID, postID)
	if err != nil {
		if err == impart.ErrNotFound {
			return impart.NewError(err, "Post not found")
		}
		return impart.NewError(err, "error deleting post")
	}

	return nil
}
