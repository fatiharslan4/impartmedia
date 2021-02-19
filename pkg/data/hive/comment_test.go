package data

import (
	"net"
	"testing"
	"time"

	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/stretchr/testify/assert"
)

var commentData = func() Comments {
	_, err := net.DialTimeout("tcp", "localhost:8000", time.Duration(1*time.Second))
	if err != nil {
		panic(err)
	}
	p, err := NewCommentData("us-east-2", localDynamo, "local", logger)
	if err != nil {
		panic(err)
	}
	return p
}()

func TestDynamo_NewComment(t *testing.T) {
	c := models.Comment{}
	c.Random()
	//c.CommentDatetime = impart.CurrentUTC()

	cc, err := commentData.NewComment(c)
	assert.NoError(t, err)
	assert.Equal(t, c, cc)

	c2 := models.Comment{}
	c2.Random()
	//c2.CommentDatetime = now()

	cc, err = commentData.NewComment(c2)
	assert.NoError(t, err)
	assert.Equal(t, c2, cc)

}

func TestDynamo_EditComment(t *testing.T) {
	c := models.Comment{}
	c.Random()
	//c.CommentDatetime = now()

	cc, err := commentData.NewComment(c)
	assert.NoError(t, err)
	assert.Equal(t, c, cc)

	newEdit := models.RandomEdit()
	c.Edits = append(c.Edits, newEdit)

	newContent := c.Content.Markdown + models.RandomContent(1).Markdown
	c.Content.Markdown = newContent

	editedComment, err := commentData.EditComment(c)
	assert.NoError(t, err)
	assert.Equal(t, c.Content.Markdown, editedComment.Content.Markdown)
	assert.Contains(t, c.Edits, newEdit)

}

func TestDynamo_GetComments(t *testing.T) {
	c := models.Comment{}
	c.Random()
	//c.CommentDatetime = now()

	cc, err := commentData.NewComment(c)
	assert.NoError(t, err)
	assert.Equal(t, c, cc)

	staticHiveID := c.HiveID
	staticPostID := c.PostID

	for i := 0; i < 9; i++ {
		c.Random()
		c.HiveID = staticHiveID
		c.PostID = staticPostID
		_, err = commentData.NewComment(c)
		assert.NoError(t, err)
	}

	posts, _, err := commentData.GetComments(staticPostID, 10, nil)
	assert.NoError(t, err)
	assert.Len(t, posts, 10)
}

func TestDynamo_GetCommentsByImpartWealthID(t *testing.T) {
	c := models.Comment{}
	c.Random()
	//c.CommentDatetime = now()

	cc, err := commentData.NewComment(c)
	assert.NoError(t, err)
	assert.Equal(t, c, cc)

	staticHiveID := c.HiveID
	staticPostID := c.PostID
	staticImpartWealthID := c.ImpartWealthID

	for i := 0; i < 9; i++ {
		c.Random()
		c.HiveID = staticHiveID
		c.PostID = staticPostID
		if i%2 == 0 && i > 0 {
			c.ImpartWealthID = staticImpartWealthID
		}
		_, err = commentData.NewComment(c)
		assert.NoError(t, err)
	}

	n := time.Time{}
	comments, err := commentData.GetCommentsByImpartWealthID(staticImpartWealthID, 10, n)
	assert.NoError(t, err)
	assert.Len(t, comments, 5)
}

func TestDynamo_DeleteComment(t *testing.T) {
	c := models.Comment{}
	c.Random()
	//c.CommentDatetime = now()

	cc, err := commentData.NewComment(c)
	assert.NoError(t, err)
	assert.Equal(t, c, cc)

	c2 := models.Comment{}
	c2.Random()
	//c2.CommentDatetime = now()

	cc, err = commentData.NewComment(c2)
	assert.NoError(t, err)
	assert.Equal(t, c2, cc)

	err = commentData.DeleteComment(c.PostID, c.CommentID)
	assert.NoError(t, err)
	err = commentData.DeleteComment(c2.PostID, c2.CommentID)
	assert.NoError(t, err)

}

func TestDynamo_GetComments_paging(t *testing.T) {
	c := models.Comment{}
	c.Random()
	//c.CommentDatetime = now()

	cc, err := commentData.NewComment(c)
	assert.NoError(t, err)
	assert.Equal(t, c, cc)

	staticHiveID := c.HiveID
	staticPostID := c.PostID

	for i := 0; i < 12; i++ {
		time.Sleep(100 * time.Millisecond)
		c.Random()
		c.HiveID = staticHiveID
		c.PostID = staticPostID
		//c.CommentDatetime = now()
		_, err = commentData.NewComment(c)
		assert.NoError(t, err)
	}

	allComments, _, err := commentData.GetComments(staticPostID, 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, 13, len(allComments))

	comments, nextPage, err := commentData.GetComments(staticPostID, 5, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(comments))
	assert.NotNil(t, nextPage)

	assert.Equal(t, allComments[0], comments[0])
	assert.Equal(t, allComments[1], comments[1])
	assert.Equal(t, allComments[2], comments[2])
	assert.Equal(t, allComments[3], comments[3])
	assert.Equal(t, allComments[4], comments[4])

	comments2, nextPage, err := commentData.GetComments(staticPostID, 0, nextPage)
	assert.NoError(t, err)
	assert.Equal(t, 8, len(comments2))
	assert.Nil(t, nextPage)
}
