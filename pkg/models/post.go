package models

import (
	"encoding/json"
	"reflect"
	"sort"
	"time"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/segmentio/ksuid"
)

type PagedPostsResponse struct {
	Posts    Posts     `json:"posts"`
	NextPage *NextPage `json:"nextPage"`
}

type Posts []Post
type Post struct {
	HiveID              string           `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	IsPinnedPost        bool             `json:"isPinnedPost"`
	PostID              string           `json:"postId"`
	PostDatetime        time.Time        `json:"postDatetime"`
	LastCommentDatetime time.Time        `json:"lastCommentDatetime"`
	Edits               Edits            `json:"edits,omitempty"`
	ImpartWealthID      string           `json:"impartWealthId"`
	ScreenName          string           `json:"screenName"`
	Subject             string           `json:"subject"`
	Content             Content          `json:"content"`
	CommentCount        int              `json:"commentCount"`
	TagIDs              tags.TagIDs      `json:"tags"`
	UpVotes             int              `json:"upVotes"`
	DownVotes           int              `json:"downVotes"`
	PostCommentTrack    PostCommentTrack `json:"postCommentTrack,omitempty"`
	Comments            Comments         `json:"comments,omitempty"`
	NextCommentPage     *NextPage        `json:"nextCommentPage"`
}

func (posts Posts) Latest() time.Time {
	var t = time.Unix(0, 0)
	for _, post := range posts {
		if post.PostDatetime.After(t) {
			t = post.PostDatetime
		}
	}
	return t
}

func (posts Posts) SortAscending(byLastComment bool) {
	sort.Slice(posts, func(i, j int) bool {
		if byLastComment {
			return posts[i].LastCommentDatetime.Before(posts[j].LastCommentDatetime)
		}
		return posts[i].PostDatetime.Before(posts[j].PostDatetime)
	})
}

func (posts Posts) SortDescending(byLastComment bool) {
	sort.Slice(posts, func(i, j int) bool {
		if byLastComment {
			return posts[i].LastCommentDatetime.After(posts[j].LastCommentDatetime)
		}
		return posts[i].PostDatetime.After(posts[j].PostDatetime)
	})
}

func (p Post) ToJson() string {
	b, _ := json.MarshalIndent(&p, "", "\t")
	return string(b)
}

func (p Post) Equals(pc Post) bool {
	return reflect.DeepEqual(p, pc)
}

func (p Post) Copy() Post {
	return p
}

func (p *Post) Random() {
	*p = Post{
		HiveID:              ksuid.New().String(),
		PostID:              ksuid.New().String(),
		Edits:               []Edit{RandomEdit()},
		ImpartWealthID:      ksuid.New().String(),
		ScreenName:          r.SillyName(),
		Subject:             r.SillyName() + r.Adjective(),
		Content:             RandomContent(7),
		CommentCount:        0,
		UpVotes:             0,
		DownVotes:           0,
		PostDatetime:        impart.CurrentUTC(),
		LastCommentDatetime: impart.CurrentUTC(),
	}
}
