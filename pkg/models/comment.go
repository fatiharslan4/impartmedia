package models

import (
	"encoding/json"
	"reflect"
	"sort"
	"time"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/leebenson/conform"
	"github.com/segmentio/ksuid"
)

type PagedCommentsResponse struct {
	Comments Comments  `json:"comments"`
	NextPage *NextPage `json:"nextPage"`
}

type Comments []Comment

type Comment struct {
	HiveID           string           `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	PostID           string           `json:"postId" jsonschema:"minLength=27,maxLength=27"`
	CommentID        string           `json:"commentId,omitempty"`
	CommentDatetime  time.Time        `json:"commentDatetime,omitempty"`
	ImpartWealthID   string           `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27"`
	ScreenName       string           `json:"screenName"`
	Content          Content          `json:"content"`
	Edits            Edits            `json:"edits,omitempty"`
	UpVotes          int              `json:"upVotes,"`
	DownVotes        int              `json:"downVotes"`
	PostCommentTrack PostCommentTrack `json:"postCommentTrack,omitempty"`
}

func (comments Comments) Latest() time.Time {
	var t = time.Unix(0, 0)
	for _, comment := range comments {
		if comment.CommentDatetime.After(t) {
			t = comment.CommentDatetime
		}
	}
	return t
}

func (comments Comments) SortAscending() {
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CommentDatetime.Before(comments[j].CommentDatetime)
	})
}

func (comments Comments) SortDescending() {
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CommentDatetime.After(comments[j].CommentDatetime)
	})
}
func (c Comment) ToJson() string {
	b, _ := json.MarshalIndent(&c, "", "\t")
	return string(b)
}

func (c Comment) Equals(cc Comment) bool {
	return reflect.DeepEqual(c, cc)
}

func (c Comment) Copy() Comment {
	return c
}

func (c *Comment) Random() {

	*c = Comment{
		HiveID:          ksuid.New().String(),
		PostID:          ksuid.New().String(),
		CommentID:       ksuid.New().String(),
		ImpartWealthID:  ksuid.New().String(),
		ScreenName:      r.SillyName(),
		Content:         RandomContent(7),
		Edits:           []Edit{RandomEdit()},
		UpVotes:         0,
		DownVotes:       0,
		CommentDatetime: impart.CurrentUTC(),
	}
	_ = conform.Strings(c)

}

func (comments Comments) ContentIDs() []string {
	out := make([]string, len(comments))

	for i, comment := range comments {
		out[i] = comment.CommentID
	}

	return out
}

func (comments Comments) AppendContentTracks(tracks map[string]PostCommentTrack) {

	for i := 0; i < len(comments); i++ {
		if currentTrack, ok := tracks[comments[i].CommentID]; ok {
			comments[i].PostCommentTrack = currentTrack
		}
	}
}
