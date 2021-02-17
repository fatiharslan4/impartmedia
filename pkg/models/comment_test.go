package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComments_AppendContentTracks(t *testing.T) {
	size := 3
	comments := make(Comments, size)

	for i := 0; i < size; i++ {
		c := Comment{}
		c.Random()
		comments[i] = c
	}

	assert.Len(t, comments, 3)
	assert.Empty(t, comments[0].PostCommentTrack)
	assert.Empty(t, comments[1].PostCommentTrack)
	assert.Empty(t, comments[2].PostCommentTrack)

	ct := map[string]PostCommentTrack{
		comments[0].CommentID: {
			ContentID: comments[0].CommentID,
			UpVoted:   true,
		},
		comments[1].CommentID: {
			ContentID: comments[1].CommentID,
			DownVoted: true,
		},
		comments[2].CommentID: {
			ContentID: comments[2].CommentID,
			Saved:     true,
		},
	}

	comments.AppendContentTracks(ct)

	assert.True(t, comments[0].PostCommentTrack.UpVoted)
	assert.True(t, comments[1].PostCommentTrack.DownVoted)
	assert.True(t, comments[2].PostCommentTrack.Saved)

}

func TestComments_Sorts(t *testing.T) {
	commentsAscending := make(Comments, 0)
	for i := 0; i < 5; i++ {
		c := Comment{}
		c.Random()
		c.CommentDatetime = time.Now()
		commentsAscending = append(commentsAscending, c)
		time.Sleep(1 * time.Second)
	}

	commentsDescending := make(Comments, 0)
	commentsResortedAscending := make(Comments, 0)
	for _, c := range commentsAscending {
		commentsDescending = append(commentsDescending, c)
		commentsResortedAscending = append(commentsResortedAscending, c)
	}
	commentsDescending.SortDescending()

	assert.Equal(t, commentsAscending[0], commentsDescending[len(commentsDescending)-1])
	assert.NotEqual(t, commentsAscending, commentsDescending)
	assert.Equal(t, commentsAscending, commentsResortedAscending)
	assert.True(t, commentsAscending[0].CommentDatetime.Before(commentsAscending[1].CommentDatetime))
}
