package models

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/leebenson/conform"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
)

type PagedCommentsResponse struct {
	Comments Comments  `json:"comments"`
	NextPage *NextPage `json:"nextPage"`
}

type Comments []Comment

type Comment struct {
	PostID              uint64           `json:"postId" jsonschema:"minLength=27,maxLength=27"`
	CommentID           uint64           `json:"commentId,omitempty"`
	CommentDatetime     time.Time        `json:"commentDatetime,omitempty"`
	ImpartWealthID      string           `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27" conform:"trim"`
	ScreenName          string           `json:"screenName" conform:"trim"`
	Content             Content          `json:"content"`
	Edits               Edits            `json:"edits,omitempty"`
	UpVotes             int              `json:"upVotes,"`
	DownVotes           int              `json:"downVotes"`
	PostCommentTrack    PostCommentTrack `json:"postCommentTrack,omitempty"`
	ReportedCount       int              `json:"reportedCount"`
	Obfuscated          bool             `json:"obfuscated"`
	Reviewed            bool             `json:"reviewed"`
	ReviewComment       string           `json:"reviewComment"`
	ReviewedDatetime    time.Time        `json:"reviewedDatetime,omitempty"`
	ParentCommentID     uint64           `json:"parentCommentId"`
	Deleted             bool             `json:"deleted,omitempty"`
	FirstName           string           `json:"firstName,omitempty"`
	LastName            string           `json:"lastName,omitempty"`
	FullName            string           `json:"fullName,omitempty"`
	AvatarBackground    string           `json:"avatarBackground,omitempty"`
	AvatarLetter        string           `json:"avatarLetter,omitempty"`
	Admin               bool             `json:"admin"`
	LoggedInUserDetails LoggedInUser     `json:"loggedInUserDetails"`
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
		PostID:          uint64(r.Number(math.MaxInt32)),
		CommentID:       uint64(r.Number(math.MaxInt32)),
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

//func (c Comment) ToDBModel() *dbmodels.Comment {
//	return &dbmodels.Comment{
//		PostID:          c.PostID,
//		ImpartWealthID:  "",
//		CreatedAt:       time.Time{},
//		Content:         "",
//		LastReplyTS:     time.Time{},
//		ParentCommentID: null.Uint64{},
//		UpVoteCount:     0,
//		DownVoteCount:   0,
//		R:               nil,
//		L:               ,
//	}
//}

func CommentsFromDBModelSlice(comments dbmodels.CommentSlice, loggedInUser *dbmodels.User) Comments {
	out := make(Comments, len(comments))
	for i, c := range comments {
		out[i] = CommentFromDBModel(c, loggedInUser)
	}
	return out
}

func CommentFromDBModel(c *dbmodels.Comment, loggedInUser *dbmodels.User) Comment {
	out := Comment{
		PostID:          c.PostID,
		CommentID:       c.CommentID,
		CommentDatetime: c.CreatedAt,
		ImpartWealthID:  c.ImpartWealthID,
		Content: Content{
			Markdown: c.Content,
		},
		//Edits:            nil,
		UpVotes:          c.UpVoteCount,
		DownVotes:        c.DownVoteCount,
		PostCommentTrack: PostCommentTrack{},
		ReportedCount:    c.ReportedCount,
		Obfuscated:       c.Obfuscated,
		Reviewed:         c.Reviewed,
		ReviewComment:    c.ReviewComment.String,
	}
	if c.ReviewedAt.Valid {
		out.ReviewedDatetime = c.ReviewedAt.Time
	}
	if c.R.ImpartWealth != nil {
		out.ScreenName = c.R.ImpartWealth.ScreenName
		out.FirstName = strings.Title(c.R.ImpartWealth.FirstName)
		out.LastName = strings.Title(c.R.ImpartWealth.LastName)
		out.FullName = strings.Title(fmt.Sprintf("%s %s", c.R.ImpartWealth.FirstName, c.R.ImpartWealth.LastName))
		out.AvatarBackground = c.R.ImpartWealth.AvatarBackground
		out.AvatarLetter = c.R.ImpartWealth.AvatarLetter
		out.Admin = c.R.ImpartWealth.Admin
	}
	if len(c.R.CommentReactions) > 0 {
		out.PostCommentTrack = PostCommentTrackFromDB(nil, c.R.CommentReactions[0])
	}

	if (c.DeletedAt != null.Time{}) {
		out.Deleted = true
	}
	// check the user is blocked
	if c.R.ImpartWealth != nil && c.R.ImpartWealth.Blocked {
		out.ScreenName = types.AccountRemoved.ToString()
	}
	if c.R.ImpartWealth == nil {
		out.ScreenName = types.AccountDeleted.ToString()
	}
	out.LoggedInUserDetails = LoggedInUser{Admin: loggedInUser.Admin,
		FirstName:        loggedInUser.FirstName,
		LastName:         loggedInUser.LastName,
		AvatarBackground: strings.Title(loggedInUser.AvatarBackground),
		AvatarLetter:     strings.Title(loggedInUser.AvatarLetter),
		FullName:         strings.Title(fmt.Sprintf("%s %s", loggedInUser.FirstName, loggedInUser.LastName)),
	}
	return out
}

type CommentNotificationInput struct {
	Ctx             context.Context
	CommentID       uint64
	ActionType      types.Type // Report,upvote,downvote, take vote
	ActionData      string
	PostID          uint64
	NotifyPostOwner bool
}

type CommentNotificationBuildDataOutput struct {
	Alert             impart.Alert
	PostOwnerAlert    impart.Alert
	PostOwnerWealthID string
}

func CommentsWithLimit(comments dbmodels.CommentSlice, limit int, loggedInUser *dbmodels.User) Comments {
	out := make(Comments, limit)
	for i, c := range comments {
		if i >= limit {
			return out
		}
		out[i] = CommentFromDBModel(c, loggedInUser)
	}
	return out
}
