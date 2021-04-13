// +build integration

package data

import (
	"github.com/impartwealthapp/backend/pkg/impart"
)

func (s *HiveTestSuite) TestPostVotes() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	//commentID := s.bootstrapComment(ctx, postID)

	err := s.hiveData.AddUpVote(ctx, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err := s.hiveData.GetPost(ctx, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//upvote twice
	err = s.hiveData.AddUpVote(ctx, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.Error(err)
	s.Equal(impart.ErrNoOp, err)
	p, err = s.hiveData.GetPost(ctx, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	ctx2 := s.contextWithImpartAdmin()
	err = s.hiveData.AddUpVote(ctx2, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(2, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//take it
	err = s.hiveData.TakeUpVote(ctx2, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//add it back
	err = s.hiveData.AddUpVote(ctx2, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(2, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//downvote instead
	err = s.hiveData.AddDownVote(ctx2, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(1, p.DownVoteCount)

	//take the downvote
	err = s.hiveData.TakeDownVote(ctx2, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//add downvote back
	err = s.hiveData.AddDownVote(ctx2, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(1, p.DownVoteCount)

	//re upvote instead
	err = s.hiveData.AddUpVote(ctx2, ContentInput{
		Type: Post,
		Id:   postID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(2, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)
}

func (s *HiveTestSuite) TestCommentVotes() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	commentID := s.bootstrapComment(ctx, postID)

	ctxUser := impart.GetCtxUser(ctx)

	err := s.hiveData.AddUpVote(ctx, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err := s.hiveData.GetComment(ctx, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	s.Equal(ctxUser.ScreenName, p.R.ImpartWealth.ScreenName)
	s.Require().Equal(1, len(p.R.CommentReactions))
	s.Require().Equal(true, p.R.CommentReactions[0].Upvoted)

	//upvote twice
	err = s.hiveData.AddUpVote(ctx, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.Error(err)
	s.Equal(impart.ErrNoOp, err)
	p, err = s.hiveData.GetComment(ctx, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	ctx2 := s.contextWithImpartAdmin()
	err = s.hiveData.AddUpVote(ctx2, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetComment(ctx2, postID)
	s.NoError(err)
	s.Equal(2, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//take it
	err = s.hiveData.TakeUpVote(ctx2, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetComment(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//add it back
	err = s.hiveData.AddUpVote(ctx2, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetComment(ctx2, postID)
	s.NoError(err)
	s.Equal(2, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//downvote instead
	err = s.hiveData.AddDownVote(ctx2, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetComment(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(1, p.DownVoteCount)

	//take the downvote
	err = s.hiveData.TakeDownVote(ctx2, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetComment(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)

	//add downvote back
	err = s.hiveData.AddDownVote(ctx2, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetComment(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.UpVoteCount)
	s.Equal(1, p.DownVoteCount)

	//re upvote instead
	err = s.hiveData.AddUpVote(ctx2, ContentInput{
		Type: Comment,
		Id:   commentID,
	})
	s.NoError(err)
	p, err = s.hiveData.GetComment(ctx2, postID)
	s.NoError(err)
	s.Equal(2, p.UpVoteCount)
	s.Equal(0, p.DownVoteCount)
}

func (s *HiveTestSuite) TestPostReports() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	//commentID := s.bootstrapComment(ctx, postID)

	p, err := s.hiveData.GetPost(ctx, postID)
	s.NoError(err)
	s.Equal(0, p.ReportedCount)
	s.Nil(p.R.PostReactions)
	s.False(p.Obfuscated)
	s.False(p.ReviewedAt.Valid)

	err = s.hiveData.ReportPost(ctx, postID, nil, false)
	s.NoError(err)

	p, err = s.hiveData.GetPost(ctx, postID)
	s.NoError(err)
	s.Equal(1, p.ReportedCount)
	s.NotNil(p.R.PostReactions)
	s.True(p.Obfuscated)
	s.False(p.ReviewedAt.Valid)
	s.False(p.R.PostReactions[0].ReportedReason.Valid)

	err = s.hiveData.ReportPost(ctx, postID, nil, false)
	s.Error(err)
	s.ErrorIs(err, impart.ErrNoOp)

	ctx2 := s.contextWithImpartAdmin()
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(1, p.ReportedCount)
	s.Nil(p.R.PostReactions)
	s.True(p.Obfuscated)
	s.False(p.ReviewedAt.Valid)

	reason := "reported"
	err = s.hiveData.ReportPost(ctx2, postID, &reason, false)
	s.NoError(err)

	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(2, p.ReportedCount)
	s.NotNil(p.R.PostReactions)
	s.True(p.Obfuscated)
	s.False(p.ReviewedAt.Valid)
	s.True(p.R.PostReactions[0].ReportedReason.Valid)
	s.Equal(reason, p.R.PostReactions[0].ReportedReason.String)

	err = s.hiveData.ReportPost(ctx, postID, nil, true)
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx, postID)
	s.NoError(err)
	s.Equal(1, p.ReportedCount)
	s.NotNil(p.R.PostReactions)
	s.True(p.Obfuscated)
	s.False(p.ReviewedAt.Valid)

	err = s.hiveData.ReportPost(ctx2, postID, nil, true)
	s.NoError(err)
	p, err = s.hiveData.GetPost(ctx2, postID)
	s.NoError(err)
	s.Equal(0, p.ReportedCount)
	s.NotNil(p.R.PostReactions)
	s.False(p.Obfuscated)
	s.False(p.ReviewedAt.Valid)

}

func (s *HiveTestSuite) TestCommentReports() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	commentID := s.bootstrapComment(ctx, postID)

	c, err := s.hiveData.GetComment(ctx, commentID)
	s.NoError(err)
	s.Equal(0, c.ReportedCount)
	s.Nil(c.R.CommentReactions)
	s.False(c.Obfuscated)
	s.False(c.ReviewedAt.Valid)

	err = s.hiveData.ReportComment(ctx, commentID, nil, false)
	s.NoError(err)

	c, err = s.hiveData.GetComment(ctx, commentID)
	s.NoError(err)
	s.Equal(1, c.ReportedCount)
	s.NotNil(c.R.CommentReactions)
	s.True(c.Obfuscated)
	s.False(c.ReviewedAt.Valid)
	s.False(c.R.CommentReactions[0].ReportedReason.Valid)

	err = s.hiveData.ReportComment(ctx, commentID, nil, false)
	s.Error(err)
	s.ErrorIs(err, impart.ErrNoOp)

	ctx2 := s.contextWithImpartAdmin()
	c, err = s.hiveData.GetComment(ctx2, commentID)
	s.NoError(err)
	s.Equal(1, c.ReportedCount)
	s.Nil(c.R.CommentReactions)
	s.True(c.Obfuscated)
	s.False(c.ReviewedAt.Valid)

	reason := "reported"
	err = s.hiveData.ReportComment(ctx2, commentID, &reason, false)
	s.NoError(err)

	c, err = s.hiveData.GetComment(ctx2, commentID)
	s.NoError(err)
	s.Equal(2, c.ReportedCount)
	s.NotNil(c.R.CommentReactions)
	s.True(c.Obfuscated)
	s.False(c.ReviewedAt.Valid)
	s.True(c.R.CommentReactions[0].ReportedReason.Valid)
	s.Equal(reason, c.R.CommentReactions[0].ReportedReason.String)

	err = s.hiveData.ReportComment(ctx, commentID, nil, true)
	s.NoError(err)
	c, err = s.hiveData.GetComment(ctx, commentID)
	s.NoError(err)
	s.Equal(1, c.ReportedCount)
	s.NotNil(c.R.CommentReactions)
	s.True(c.Obfuscated)
	s.False(c.ReviewedAt.Valid)

	err = s.hiveData.ReportComment(ctx2, commentID, nil, true)
	s.NoError(err)
	c, err = s.hiveData.GetComment(ctx2, commentID)
	s.NoError(err)
	s.Equal(0, c.ReportedCount)
	s.NotNil(c.R.CommentReactions)
	s.False(c.Obfuscated)
	s.False(c.ReviewedAt.Valid)

}
