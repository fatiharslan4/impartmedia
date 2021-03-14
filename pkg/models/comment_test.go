package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
