// +build integration

package data

import (
	"context"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
)

func (s *HiveTestSuite) bootstrapComment(ctx context.Context, postId uint64) uint64 {
	var err error
	ctxUser := impart.GetCtxUser(ctx)
	newComment := &dbmodels.Comment{
		PostID:          postId,
		ImpartWealthID:  ctxUser.ImpartWealthID,
		CreatedAt:       impart.CurrentUTC(),
		Content:         "moleh moleh moleh",
		LastReplyTS:     impart.CurrentUTC(),
		ParentCommentID: null.Uint64{},
	}
	newComment, err = s.hiveData.NewComment(ctx, newComment)
	s.NoError(err)
	s.NotZero(newComment.CommentID)
	return newComment.CommentID
}

func (s *HiveTestSuite) TestGetNewComment() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	ctxUser := impart.GetCtxUser(ctx)
	postID := s.bootstrapPost(ctx, hiveID)

	newComment := dbmodels.Comment{
		PostID:          postID,
		ImpartWealthID:  ctxUser.ImpartWealthID,
		CreatedAt:       impart.CurrentUTC(),
		Content:         "moleh moleh moleh",
		LastReplyTS:     impart.CurrentUTC(),
		ParentCommentID: null.Uint64{},
	}
	tmp := newComment
	createdComment, err := s.hiveData.NewComment(ctx, &tmp)
	s.NoError(err)
	s.NotZero(createdComment.CommentID)
	s.Equal(newComment.PostID, createdComment.PostID)
	s.Equal(newComment.ImpartWealthID, createdComment.ImpartWealthID)
	s.Equal(newComment.CreatedAt, createdComment.CreatedAt)
	s.Equal(newComment.Content, createdComment.Content)
	s.Equal(newComment.LastReplyTS, createdComment.LastReplyTS)

	//comment, err
}

func (s *HiveTestSuite) TestDeleteComment() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	commentID := s.bootstrapComment(ctx, postID)

	existingComment, err := s.hiveData.GetComment(ctx, commentID)
	s.NoError(err)
	s.NotNil(existingComment)

	err = s.hiveData.DeleteComment(ctx, commentID)

	existingComment, err = s.hiveData.GetComment(ctx, commentID)
	s.NotNil(err)
	s.Equal(impart.ErrNotFound, err)
	s.Nil(existingComment)

}

func (s *HiveTestSuite) TestGetEditComment() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	commentID := s.bootstrapComment(ctx, postID)

	existingComment, err := s.hiveData.GetComment(ctx, commentID)
	s.NoError(err)

	existingComment.Content = "edited content"
	c, err := s.hiveData.EditComment(ctx, existingComment)
	s.NoError(err)
	s.Equal("edited content", c.Content)

}

func (s *HiveTestSuite) TestGetCommentsPaging() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	//ctxUser := impart.GetCtxUser(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	expectedCommentIds := []uint64{
		s.bootstrapComment(ctx, postID),
		s.bootstrapComment(ctx, postID),
		s.bootstrapComment(ctx, postID),
		s.bootstrapComment(ctx, postID),
		s.bootstrapComment(ctx, postID),
	}

	comments, nextPage, err := s.hiveData.GetComments(ctx, postID, 2, 0)
	s.NoError(err)
	s.Len(comments, 2)
	s.Require().NotNil(nextPage)
	s.Equal(2, nextPage.Offset)
	s.Equal(expectedCommentIds[0], comments[0].CommentID)
	s.Equal(expectedCommentIds[1], comments[1].CommentID)

	comments, nextPage, err = s.hiveData.GetComments(ctx, postID, 2, nextPage.Offset)
	s.NoError(err)
	s.Len(comments, 2)
	s.Require().NotNil(nextPage)
	s.Equal(4, nextPage.Offset)
	s.Equal(expectedCommentIds[2], comments[0].CommentID)
	s.Equal(expectedCommentIds[3], comments[1].CommentID)

	comments, nextPage, err = s.hiveData.GetComments(ctx, postID, 2, nextPage.Offset)
	s.NoError(err)
	s.Len(comments, 1)
	s.Require().Nil(nextPage)
	s.Equal(expectedCommentIds[4], comments[0].CommentID)
}
