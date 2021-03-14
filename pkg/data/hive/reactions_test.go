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
