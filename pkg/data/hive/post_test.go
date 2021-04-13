// +build integration

package data

import (
	"context"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/segmentio/ksuid"
	"sort"
)

func (s *HiveTestSuite) bootstrapPost(ctx context.Context, hiveId uint64) uint64 {
	ctxUser := impart.GetCtxUser(ctx)
	post := &dbmodels.Post{
		HiveID:         hiveId,
		ImpartWealthID: ctxUser.ImpartWealthID,
		Pinned:         false,
		CreatedAt:      impart.CurrentUTC(),
		Subject:        "subject",
		Content:        "some content",
		LastCommentTS:  impart.CurrentUTC(),
		CommentCount:   0,
		UpVoteCount:    0,
		DownVoteCount:  0,
	}
	post, err := s.hiveData.NewPost(ctx, post, dbmodels.TagSlice{})
	s.NoError(err)
	s.NotZero(post.PostID)
	return post.PostID
}
func (s *HiveTestSuite) TestCreatePost() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	ctxUser := impart.GetCtxUser(ctx)

	post := &dbmodels.Post{
		HiveID:         hiveID,
		ImpartWealthID: ctxUser.ImpartWealthID,
		Pinned:         false,
		CreatedAt:      impart.CurrentUTC(),
		Subject:        "subject",
		Content:        "some content",
		LastCommentTS:  impart.CurrentUTC(),
		CommentCount:   0,
		UpVoteCount:    0,
		DownVoteCount:  0,
	}

	post, err := s.hiveData.NewPost(ctx, post, dbmodels.TagSlice{})
	s.NoError(err)
	s.NotZero(post.PostID)

	availableTags, err := dbmodels.Tags().All(ctx, s.db)
	s.NoError(err)

	err = post.AddTags(ctx, s.db, false, &dbmodels.Tag{TagID: availableTags[3].TagID})
	s.NoError(err)

	post, err = s.hiveData.GetPost(ctx, post.PostID)
	s.NoError(err)
	s.Equal(post.PostID, post.PostID)
	s.Equal(availableTags[3].TagID, post.R.Tags[0].TagID)
}

func (s *HiveTestSuite) TestEditPost() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	//ctxUser := impart.GetCtxUser(ctx)
	postID := s.bootstrapPost(ctx, hiveID)

	post, err := s.hiveData.GetPost(ctx, postID)
	s.NoError(err)
	s.Equal(postID, post.PostID)

	newContent := "new content" + ksuid.New().String()
	post.Content = newContent
	post, err = s.hiveData.EditPost(ctx, post, dbmodels.TagSlice{})
	s.NoError(err)
	s.Equal(post.Content, newContent)
	s.Equal(postID, post.PostID)
}

func (s *HiveTestSuite) TestDeletePost() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	//ctxUser := impart.GetCtxUser(ctx)
	postID := s.bootstrapPost(ctx, hiveID)

	err := s.hiveData.DeletePost(ctx, postID)
	s.NoError(err)

	_, err = s.hiveData.GetPost(ctx, postID)
	s.EqualError(err, impart.ErrNotFound.Error())
}

func (s *HiveTestSuite) TestGetPostsPaging() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	ctxUser := impart.GetCtxUser(ctx)
	expectedPostIds := []uint64{
		s.bootstrapPost(ctx, hiveID),
		s.bootstrapPost(ctx, hiveID),
		s.bootstrapPost(ctx, hiveID),
		s.bootstrapPost(ctx, hiveID),
		s.bootstrapPost(ctx, hiveID),
	}
	sort.Slice(expectedPostIds, func(i, j int) bool {
		return j < i
	})
	gpi := GetPostsInput{
		HiveID:              hiveID,
		Limit:               2,
		Offset:              0,
		IsLastCommentSorted: false,
		TagIDs:              nil,
	}

	posts, nextPage, err := s.hiveData.GetPosts(ctx, gpi)
	s.NoError(err)
	s.Len(posts, 2)
	s.Require().NotNil(nextPage)
	s.Equal(2, nextPage.Offset)
	s.Equal(expectedPostIds[0], posts[0].PostID)
	s.Equal(expectedPostIds[1], posts[1].PostID)
	s.Equal(ctxUser.ScreenName, posts[0].R.ImpartWealth.ScreenName)

	gpi.Offset = nextPage.Offset
	posts, nextPage, err = s.hiveData.GetPosts(ctx, gpi)
	s.NoError(err)
	s.Len(posts, 2)
	s.Require().NotNil(nextPage)
	s.Equal(4, nextPage.Offset)
	s.Equal(expectedPostIds[2], posts[0].PostID)
	s.Equal(expectedPostIds[3], posts[1].PostID)

	gpi.Offset = nextPage.Offset
	posts, nextPage, err = s.hiveData.GetPosts(ctx, gpi)
	s.NoError(err)
	s.Len(posts, 1)
	s.Require().Nil(nextPage)
	s.Equal(expectedPostIds[4], posts[0].PostID)
}
