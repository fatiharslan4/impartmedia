package data

import (
	"net"
	"testing"
	"time"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

//var postData, _ = NewPostData("us-east-2", localDynamo, "local", logger)

var postData = func() Posts {
	_, err := net.DialTimeout("tcp", "localhost:8000", time.Duration(1*time.Second))
	if err != nil {
		panic(err)
	}
	p, err := NewPostData("us-east-2", localDynamo, "local", logger)
	if err != nil {
		panic(err)
	}
	return p
}()

//var rawDynamo = func() *dynamo {
//	_, err := net.DialTimeout("tcp", "localhost:8000", time.Duration(1*time.Second))
//	if err != nil {
//		panic(err)
//	}
//	p, err := newDynamo("us-east-2", localDynamo, "local", logger)
//	if err != nil {
//		panic(err)
//	}
//	return p
//}()

func TestNewPostData(t *testing.T) {
	p := models.Post{}
	p.Random()
	ts := impart.CurrentUTC()
	p.LastCommentDatetime = ts
	p.PostDatetime = ts
	cp, err := postData.NewPost(p)
	assert.NoError(t, err)
	assert.Equal(t, p, cp)

	p2 := models.Post{}
	p2.Random()
	p2.HiveID = p.HiveID

	cp, err = postData.NewPost(p2)
	assert.NoError(t, err)
	assert.Equal(t, p2, cp)
}

func TestDynamo_EditPost(t *testing.T) {
	p := models.Post{}
	p.Random()

	cp, err := postData.NewPost(p)
	assert.NoError(t, err)
	assert.Equal(t, p.Content, cp.Content)

	newEdit := models.RandomEdit()
	p.Edits = append(p.Edits, newEdit)

	newContent := p.Content.Markdown + models.RandomContent(1).Markdown
	p.Content.Markdown = newContent

	cp, err = postData.EditPost(p)
	assert.NoError(t, err)
	assert.Equal(t, p.Content.Markdown, cp.Content.Markdown)
	assert.Contains(t, cp.Edits, newEdit)
	assert.Equal(t, p.Content, cp.Content)
}

func TestDynamo_GetPosts(t *testing.T) {
	p := models.Post{}
	p.Random()
	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	staticHiveID := p.HiveID

	for i := 0; i < 9; i++ {
		p.Random()
		p.HiveID = staticHiveID
		_, err = postData.NewPost(p)
		assert.NoError(t, err)
	}

	posts, _, err := postData.GetPosts(GetPostsInput{HiveID: staticHiveID})
	assert.NoError(t, err)
	assert.Len(t, posts, 10)
}

func TestDynamo_GetPostsPaging(t *testing.T) {
	p := models.Post{}
	p.Random()
	//p.PostDatetime = now()
	//p.LastCommentDatetime = now()
	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	staticHiveID := p.HiveID

	for i := 0; i < 9; i++ {
		time.Sleep(time.Millisecond * 10)
		p.Random()
		//p.PostDatetime = now()
		//p.LastCommentDatetime = now()
		p.HiveID = staticHiveID
		_, err = postData.NewPost(p)
		assert.NoError(t, err)
	}

	posts, _, err := postData.GetPosts(GetPostsInput{HiveID: staticHiveID, Limit: 7})
	assert.NoError(t, err)
	assert.Len(t, posts, 7)
	priorPostDatetime := impart.CurrentUTC().Add(time.Second)
	for i := 0; i < 7; i++ {
		//ensure descending sort
		currPostDateTime := posts[i].PostDatetime
		assert.True(t, priorPostDatetime.After(currPostDateTime))
		priorPostDatetime = currPostDateTime
		assert.Equal(t, staticHiveID, posts[i].HiveID)
	}
}

func TestDynamo_GetPostsImpartWealthID(t *testing.T) {
	p := models.Post{}
	p.Random()
	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	staticImpartWealthID := p.ImpartWealthID

	for i := 0; i < 9; i++ {
		p.Random()
		p.ImpartWealthID = staticImpartWealthID
		_, err = postData.NewPost(p)
		assert.NoError(t, err)
	}

	posts, err := postData.GetPostsImpartWealthID(staticImpartWealthID, 2, DefaultTime)
	assert.NoError(t, err)
	assert.Len(t, posts, 2)
}

func TestDynamo_GetPostsByTags(t *testing.T) {
	p := models.Post{}
	p.Random()
	p.TagIDs = append(p.TagIDs, tags.Income().ID)
	_, err := postData.NewPost(p)
	assert.NoError(t, err)

	staticHiveID := p.HiveID

	for i := 0; i < 9; i++ {
		p.Random()
		p.HiveID = staticHiveID
		_, err = postData.NewPost(p)
		assert.NoError(t, err)
	}

	posts, _, err := postData.GetPosts(GetPostsInput{HiveID: staticHiveID, TagIDs: []int{1, 2}})
	assert.NoError(t, err)
	assert.Len(t, posts, 1)
}

func TestDynamo_IncrementDecrement(t *testing.T) {
	p := models.Post{}
	p.Random()
	p.CommentCount = 0
	p.UpVotes = 0
	p.DownVotes = 0

	cp, err := postData.NewPost(p)
	assert.NoError(t, err)
	assert.Equal(t, p.Content, cp.Content)
	assert.Equal(t, cp.CommentCount, 0)
	assert.Equal(t, cp.UpVotes, 0)
	assert.Equal(t, cp.DownVotes, 0)

	err = postData.IncrementDecrementPost(p.HiveID, p.PostID, CommentCountColumnName, false)
	assert.NoError(t, err)
	err = postData.IncrementDecrementPost(p.HiveID, p.PostID, UpVoteCountColumnName, false)
	assert.NoError(t, err)
	err = postData.IncrementDecrementPost(p.HiveID, p.PostID, DownVoteCountColumnName, false)
	assert.NoError(t, err)

	cp, err = postData.GetPost(p.HiveID, p.PostID, true)
	assert.NoError(t, err)
	assert.Equal(t, cp.CommentCount, 1)
	assert.Equal(t, cp.UpVotes, 1)
	assert.Equal(t, cp.DownVotes, 1)

	// increment to 2
	err = postData.IncrementDecrementPost(p.HiveID, p.PostID, CommentCountColumnName, false)
	assert.NoError(t, err)
	cp, err = postData.GetPost(p.HiveID, p.PostID, true)
	assert.NoError(t, err)
	assert.Equal(t, cp.CommentCount, 2)

	// decrement to 1
	err = postData.IncrementDecrementPost(p.HiveID, p.PostID, CommentCountColumnName, true)
	assert.NoError(t, err)
	cp, err = postData.GetPost(p.HiveID, p.PostID, true)
	assert.NoError(t, err)
	assert.Equal(t, cp.CommentCount, 1)

	// decrement to 0
	err = postData.IncrementDecrementPost(p.HiveID, p.PostID, CommentCountColumnName, true)
	assert.NoError(t, err)
	cp, err = postData.GetPost(p.HiveID, p.PostID, true)
	assert.NoError(t, err)
	assert.Equal(t, cp.CommentCount, 0)

	// attempt to decrement past zero; should be no error, should be only zero.
	err = postData.IncrementDecrementPost(p.HiveID, p.PostID, CommentCountColumnName, true)
	assert.NoError(t, err)
	cp, err = postData.GetPost(p.HiveID, p.PostID, true)
	assert.NoError(t, err)
	assert.Equal(t, cp.CommentCount, 0)
}

func TestDynamo_UpdateValue(t *testing.T) {
	p := models.Post{}
	p.Random()

	cp, err := postData.NewPost(p)
	assert.NoError(t, err)
	assert.Equal(t, p.LastCommentDatetime, p.LastCommentDatetime)

	futureDate := cp.LastCommentDatetime.Add(time.Hour)
	cp.LastCommentDatetime = futureDate

	err = postData.UpdateTimestampLater(cp.HiveID, cp.PostID, "lastCommentDatetime", futureDate)
	assert.NoError(t, err)

	cp, err = postData.GetPost(cp.HiveID, cp.PostID, true)
	assert.NoError(t, err)

	assert.Equal(t, futureDate, cp.LastCommentDatetime)

	priorDate := cp.LastCommentDatetime.Add(-time.Hour)
	cp.LastCommentDatetime = priorDate

	err = postData.UpdateTimestampLater(cp.HiveID, cp.PostID, "lastCommentDatetime", futureDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ConditionalCheckFailedException")

}

func TestPostPaging(t *testing.T) {
	var p models.Post
	var err error
	p.Random()
	staticHiveId := p.HiveID
	ts := impart.CurrentUTC()
	p.PostDatetime = ts
	p.LastCommentDatetime = ts

	for i := 0; i < 100; i++ {
		toCreate := models.Post{
			ImpartWealthID:      ksuid.New().String(),
			ScreenName:          p.ScreenName,
			HiveID:              staticHiveId,
			PostID:              ksuid.New().String(),
			LastCommentDatetime: ts.Add(time.Millisecond * 5 * time.Duration(i+1)),
			PostDatetime:        ts.Add(time.Millisecond * 5 * time.Duration(i+1)),
		}
		_, err = postData.NewPost(toCreate)
		assert.NoError(t, err)
	}

	allPosts, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID: staticHiveId,
		Limit:  101, //dynamodb local returns last evaluated key if total items == limit; real dynamo does not.
	})

	assert.NoError(t, err)
	assert.Len(t, allPosts, 100)
	assert.Nil(t, nextPage)

	fivePosts, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID: staticHiveId,
		Limit:  5,
	})

	assert.NoError(t, err)
	assert.Len(t, fivePosts, 5)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[:5], fivePosts)

	fivePosts, nextPage, err = postData.GetPosts(GetPostsInput{
		HiveID:   staticHiveId,
		Limit:    5,
		NextPage: nextPage,
	})

	assert.NoError(t, err)
	assert.Len(t, fivePosts, 5)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[5:10], fivePosts)

	fivePosts, nextPage, err = postData.GetPosts(GetPostsInput{
		HiveID:   staticHiveId,
		Limit:    5,
		NextPage: nextPage,
	})

	assert.NoError(t, err)
	assert.Len(t, fivePosts, 5)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[10:15], fivePosts)

	noLimit25Posts, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID:   staticHiveId,
		NextPage: nextPage,
	})

	assert.NoError(t, err)
	assert.Len(t, noLimit25Posts, 25)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[15:40], noLimit25Posts)

	remainingPosts, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID:   staticHiveId,
		NextPage: nextPage,
		Limit:    100,
	})

	assert.NoError(t, err)
	assert.Len(t, remainingPosts, 60)
	assert.Nil(t, nextPage)
	assert.Equal(t, allPosts[40:100], remainingPosts)

}

func TestPostPagingCommentSort(t *testing.T) {
	var p models.Post
	var err error
	p.Random()
	staticHiveId := p.HiveID

	for i := 0; i < 20; i++ {
		_, err = postData.NewPost(p)
		assert.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
		p.Random()
		p.HiveID = staticHiveId
		p.LastCommentDatetime = impart.CurrentUTC()
	}

	allPosts, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID:              staticHiveId,
		Limit:               50,
		IsLastCommentSorted: true,
	})

	assert.NoError(t, err)
	assert.Len(t, allPosts, 20)
	assert.Nil(t, nextPage)

	fivePosts, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID:              staticHiveId,
		Limit:               5,
		IsLastCommentSorted: true,
	})

	assert.NoError(t, err)
	assert.Len(t, fivePosts, 5)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[:5], fivePosts)

	fivePosts, nextPage, err = postData.GetPosts(GetPostsInput{
		HiveID:              staticHiveId,
		Limit:               5,
		NextPage:            nextPage,
		IsLastCommentSorted: true,
	})

	assert.NoError(t, err)
	assert.Len(t, fivePosts, 5)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[5:10], fivePosts)

	fivePosts, nextPage, err = postData.GetPosts(GetPostsInput{
		HiveID:              staticHiveId,
		Limit:               5,
		NextPage:            nextPage,
		IsLastCommentSorted: true,
	})

	assert.NoError(t, err)
	assert.Len(t, fivePosts, 5)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[10:15], fivePosts)

	fivePosts, nextPage, err = postData.GetPosts(GetPostsInput{
		HiveID:              staticHiveId,
		NextPage:            nextPage,
		IsLastCommentSorted: true,
	})

	assert.NoError(t, err)
	assert.Len(t, fivePosts, 5)
	assert.Nil(t, nextPage)
	assert.Equal(t, allPosts[15:20], fivePosts)

}

func TestPostPagingSmallLimitWithTagsFilter(t *testing.T) {
	var p models.Post
	var err error
	p.Random()
	staticHiveId := p.HiveID
	ts := impart.CurrentUTC()
	p.PostDatetime = ts
	p.LastCommentDatetime = ts

	for i := 0; i < 12; i++ {
		var tagIDs []int
		if i%2 == 0 {
			tagIDs = []int{tags.RetirementSavingsID}
		}
		toCreate := models.Post{
			ImpartWealthID:      ksuid.New().String(),
			ScreenName:          p.ScreenName,
			HiveID:              staticHiveId,
			PostID:              ksuid.New().String(),
			LastCommentDatetime: ts.Add(time.Millisecond * 5 * time.Duration(i+1)),
			PostDatetime:        ts.Add(time.Millisecond * 5 * time.Duration(i+1)),
			TagIDs:              tagIDs,
		}
		_, err = postData.NewPost(toCreate)
		assert.NoError(t, err)
	}

	allPosts, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID: staticHiveId,
		Limit:  12,
		TagIDs: []int{tags.RetirementSavingsID},
	})

	assert.NoError(t, err)
	assert.Len(t, allPosts, 6, "length of allPosts (%v) was not %v", len(allPosts), 6)
	assert.Nil(t, nextPage)

	page, nextPage, err := postData.GetPosts(GetPostsInput{
		HiveID: staticHiveId,
		Limit:  2,
		TagIDs: []int{tags.RetirementSavingsID},
	})

	assert.NoError(t, err)
	assert.Len(t, page, 2, "length of page (%v) does not match %v", len(page), 2)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[:2], page)

	page, nextPage, err = postData.GetPosts(GetPostsInput{
		HiveID:   staticHiveId,
		Limit:    2,
		TagIDs:   []int{tags.RetirementSavingsID},
		NextPage: nextPage,
	})

	assert.NoError(t, err)
	assert.Len(t, page, 2, "length of page (%v) does not match %v", len(page), 2)
	assert.NotNil(t, nextPage)
	assert.Equal(t, allPosts[2:4], page)

	page, nextPage, err = postData.GetPosts(GetPostsInput{
		HiveID:   staticHiveId,
		Limit:    2,
		TagIDs:   []int{tags.RetirementSavingsID},
		NextPage: nextPage,
	})

	assert.NoError(t, err)
	assert.Len(t, page, 2, "length of page (%v) does not match %v", len(page), 2)
	assert.Nil(t, nextPage)
	assert.Equal(t, allPosts[4:6], page)

}
