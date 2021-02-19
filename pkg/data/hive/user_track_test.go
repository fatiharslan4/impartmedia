package data

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

var trackData = func() UserTrack {
	_, err := net.DialTimeout("tcp", "localhost:8000", time.Duration(1*time.Second))
	if err != nil {
		panic(err)
	}
	t, err := NewContentTrack("us-east-2", localDynamo, "local", logger)
	if err != nil {
		panic(err)
	}
	return t
}()

func TestDynamo_GetUserTrackForContent(t *testing.T) {

	var p models.Post
	p.Random()
	var c models.Comment

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	impartWealthID := p.ImpartWealthID
	HiveID := p.HiveID
	PostID := p.PostID

	contentIDs := make([]string, 10)
	now := time.Now()

	for i := 0; i < 10; i++ {
		c.Random()
		c.HiveID = HiveID
		c.PostID = PostID
		c.ImpartWealthID = impartWealthID
		contentIDs[i] = c.CommentID

		_, err = commentData.NewComment(c)
		assert.NoError(t, err)

		err = trackData.AddUpVote(impartWealthID, c.CommentID, HiveID, PostID)
		assert.NoError(t, err)
	}
	fmt.Printf("elapsed for creation %v\n", time.Since(now))

	now = time.Now()
	trackedContent, err := trackData.GetUserTrackForContent(impartWealthID, contentIDs)
	fmt.Printf("elapsed for retrieval%v\n", time.Since(now))
	assert.NoError(t, err)
	assert.Len(t, trackedContent, 10)
}

func TestDynamo_GetContentTrack(t *testing.T) {

	var p models.Post
	p.Random()

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	contentID := p.PostID
	numTracks := 10
	now := time.Now()
	for i := 0; i < numTracks; i++ {
		err = trackData.AddUpVote(ksuid.New().String(), contentID, p.HiveID, p.PostID)
		assert.NoError(t, err)
	}
	fmt.Printf("tok %v to create %v tracks \n", time.Since(now), numTracks)

	now = time.Now()
	c, err := trackData.GetContentTrack(contentID, 0, "")
	fmt.Printf("elapsed for retrieval%v\n", time.Since(now))
	assert.NoError(t, err)
	assert.Len(t, c, numTracks)
}

func TestDynamo_TrackMisc(t *testing.T) {
	var p models.Post
	p.Random()

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	tr := models.PostCommentTrack{
		ImpartWealthID: p.ImpartWealthID,
		HiveID:         p.HiveID,
		PostID:         p.PostID,
		ContentID:      p.PostID,
	}

	err = trackData.AddUpVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err := trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.True(t, v.UpVoted)
	assert.False(t, v.DownVoted)
	assert.True(t, time.Time{}.Before(v.VotedDatetime))

	err = trackData.AddDownVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v2, err := trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.False(t, v2.UpVoted)
	assert.True(t, v2.DownVoted)
	assert.True(t, v.VotedDatetime.Before(v2.VotedDatetime))

	err = trackData.Save(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v3, err := trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.False(t, v2.UpVoted)
	assert.True(t, v2.DownVoted)
	assert.True(t, v.VotedDatetime.Before(v2.VotedDatetime))
	assert.Equal(t, v3.VotedDatetime, v2.VotedDatetime)
	assert.True(t, v3.Saved)

}
func TestDynamo_Upvote(t *testing.T) {
	var p models.Post
	p.Random()

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	tr := models.PostCommentTrack{
		ImpartWealthID: p.ImpartWealthID,
		HiveID:         p.HiveID,
		PostID:         p.PostID,
		ContentID:      p.PostID,
	}

	err = trackData.AddUpVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err := trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.True(t, v.UpVoted)
	assert.False(t, v.DownVoted)
	assert.True(t, time.Time{}.Before(v.VotedDatetime))

}

func TestDynamo_Downvote(t *testing.T) {
	var p models.Post
	p.Random()

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	tr := models.PostCommentTrack{
		ImpartWealthID: p.ImpartWealthID,
		HiveID:         p.HiveID,
		PostID:         p.PostID,
		ContentID:      p.PostID,
	}

	err = trackData.AddDownVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err := trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.True(t, v.DownVoted)
	assert.False(t, v.UpVoted)
	assert.True(t, time.Time{}.Before(v.VotedDatetime))

}

func TestDynamo_ClearVote(t *testing.T) {
	var p models.Post
	p.Random()

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	tr := models.PostCommentTrack{
		ImpartWealthID: p.ImpartWealthID,
		HiveID:         p.HiveID,
		PostID:         p.PostID,
		ContentID:      p.PostID,
	}

	err = trackData.AddDownVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err := trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.True(t, v.DownVoted)
	assert.False(t, v.UpVoted)
	assert.True(t, time.Time{}.Before(v.VotedDatetime))

	err = trackData.AddDownVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.Equal(t, impart.ErrNoOp, err)

	err = trackData.TakeDownVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err = trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.False(t, v.DownVoted)
	assert.False(t, v.UpVoted)
	assert.True(t, time.Time{}.Equal(v.VotedDatetime))

	tr = models.PostCommentTrack{
		ImpartWealthID: p.ImpartWealthID,
		HiveID:         p.HiveID,
		PostID:         p.PostID,
		ContentID:      p.PostID,
	}

	err = trackData.AddUpVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err = trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.True(t, v.UpVoted)
	assert.False(t, v.DownVoted)
	assert.True(t, time.Time{}.Before(v.VotedDatetime))

	err = trackData.AddUpVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.Equal(t, impart.ErrNoOp, err)

	err = trackData.TakeUpVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err = trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.False(t, v.DownVoted)
	assert.False(t, v.UpVoted)
	assert.True(t, time.Time{}.Equal(v.VotedDatetime))

}

func TestDynamo_CommentUpvote(t *testing.T) {
	var p models.Post
	p.Random()

	var c models.Comment
	c.Random()
	c.HiveID = p.HiveID
	c.PostID = p.PostID
	c.ImpartWealthID = p.ImpartWealthID

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	_, err = commentData.NewComment(c)
	assert.NoError(t, err)

	tr := models.PostCommentTrack{
		ImpartWealthID: p.ImpartWealthID,
		HiveID:         p.HiveID,
		PostID:         p.PostID,
		ContentID:      c.CommentID,
	}

	err = trackData.AddUpVote(tr.ImpartWealthID, tr.ContentID, tr.HiveID, tr.PostID)
	assert.NoError(t, err)

	v, err := trackData.GetUserTrack(tr.ImpartWealthID, tr.ContentID, true)
	assert.NoError(t, err)
	assert.True(t, v.UpVoted)
	assert.False(t, v.DownVoted)
	assert.True(t, time.Time{}.Before(v.VotedDatetime))

}

// This test runs against prod dev dynamoDB, rather than dynamodb local.
func TestDynamo_DeleteTracks(t *testing.T) {
	var p models.Post
	p.Random()

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	impartWealthID := p.ImpartWealthID
	HiveID := p.HiveID
	PostID := p.PostID

	numComments := 10
	contentIDs := make([]string, numComments)
	now := time.Now()
	var eg errgroup.Group
	for i := 0; i < numComments; i++ {
		idx := i
		eg.Go(func() error {
			c := models.Comment{}
			c.Random()
			c.HiveID = HiveID
			c.PostID = PostID
			c.ImpartWealthID = impartWealthID
			contentIDs[idx] = c.CommentID

			_, err = commentData.NewComment(c)
			if err != nil {
				return err
			}

			err = trackData.AddUpVote(impartWealthID, c.CommentID, HiveID, PostID)
			if err != nil {
				return err
			}

			return trackData.AddUpVote(ksuid.New().String(), PostID, HiveID, PostID)
		})
	}
	eg.Wait()
	fmt.Printf("took %v to create %v comments concurrently \n", time.Since(now), numComments)

	now = time.Now()
	trackedContent, err := trackData.GetUserTrackForContent(impartWealthID, contentIDs)
	assert.NoError(t, err)
	assert.Len(t, trackedContent, numComments)

	postTracks, err := trackData.GetContentTrack(PostID, 0, "")
	assert.NoError(t, err)
	assert.Len(t, postTracks, numComments)

	//cleanup
	now = time.Now()
	err = postData.DeletePost(p.HiveID, p.PostID)
	fmt.Printf("complete post cleanup took %v \n", time.Since(now))
	assert.NoError(t, err)

	// Verify cleaned up properly with no content tracks for the post
	postTracks, err = trackData.GetContentTrack(PostID, 0, "")
	assert.NoError(t, err)
	assert.Equal(t, len(postTracks), 0)

	// And no comment tracks for the comments
	for _, cid := range contentIDs {
		commentTracks, err := trackData.GetContentTrack(cid, 0, "")
		assert.NoError(t, err)
		assert.Equal(t, len(commentTracks), 0)
	}

	//and no comments
	comments, _, err := commentData.GetComments(PostID, 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(comments), 0)

	// and no post
	_, err = postData.GetPost(HiveID, PostID, true)
	assert.Equal(t, err, impart.ErrNotFound)

}

// This test runs against prod dev dynamoDB, rather than dynamodb local.
func TestDynamo_DeleteTrackPages(t *testing.T) {
	var p models.Post
	p.Random()
	//var c models.Comment

	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	impartWealthID := p.ImpartWealthID
	HiveID := p.HiveID
	PostID := p.PostID

	numTracks := 27
	contentIDs := make([]string, numTracks)
	now := time.Now()
	var eg errgroup.Group
	for i := 0; i < numTracks; i++ {
		idx := i
		eg.Go(func() error {
			c := models.Comment{}
			c.Random()
			c.HiveID = HiveID
			c.PostID = PostID
			c.ImpartWealthID = impartWealthID
			contentIDs[idx] = c.CommentID

			_, err = commentData.NewComment(c)
			if err != nil {
				return err
			}
			err = trackData.AddUpVote(impartWealthID, c.CommentID, HiveID, PostID)
			if err != nil {
				return err
			}

			return trackData.AddUpVote(ksuid.New().String(), PostID, HiveID, PostID)
		})
	}
	err = eg.Wait()
	assert.NoError(t, err)
	fmt.Printf("took %v to create %v comments concurrently \n", time.Since(now), numTracks)

	now = time.Now()
	trackedContent, err := trackData.GetUserTrackForContent(impartWealthID, contentIDs)
	assert.NoError(t, err)
	assert.Len(t, trackedContent, numTracks)

	postTracks, err := trackData.GetContentTrack(PostID, int64(numTracks), "")
	assert.NoError(t, err)
	assert.Len(t, postTracks, numTracks)

	//cleanup
	now = time.Now()
	err = postData.DeletePost(p.HiveID, p.PostID)
	fmt.Printf("complete post cleanup took %v \n", time.Since(now))
	assert.NoError(t, err)

	// Verify cleaned up properly with no content tracks for the post
	postTracks, err = trackData.GetContentTrack(PostID, 0, "")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(postTracks))

	// And no comment tracks for the comments
	for _, cid := range contentIDs {
		commentTracks, err := trackData.GetContentTrack(cid, 0, "")
		assert.NoError(t, err)
		assert.Equal(t, 0, len(commentTracks))
	}

	//and no comments
	comments, _, err := commentData.GetComments(PostID, 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(comments))

	// and no post
	_, err = postData.GetPost(HiveID, PostID, true)
	assert.Equal(t, err, impart.ErrNotFound)

}
