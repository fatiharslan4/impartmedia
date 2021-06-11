package models

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"sort"
	"time"

	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/segmentio/ksuid"
)

type PagedPostsResponse struct {
	Posts    Posts     `json:"posts"`
	NextPage *NextPage `json:"nextPage"`
}

type PagedReportedContentResponse struct {
	Posts    Posts     `json:"posts"`
	Comments Comments  `json:"comments"`
	NextPage *NextPage `json:"nextPage"`
}

type ReportedUser struct {
	ImpartWealthID string `json:"impartWealthId"`
	ScreenName     string `json:"screenName"`
}

type Posts []Post
type Post struct {
	HiveID              uint64           `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	IsPinnedPost        bool             `json:"isPinnedPost"`
	PostID              uint64           `json:"postId"`
	PostDatetime        time.Time        `json:"postDatetime"`
	LastCommentDatetime time.Time        `json:"lastCommentDatetime"`
	Edits               Edits            `json:"edits,omitempty"`
	ImpartWealthID      string           `json:"impartWealthId" conform:"trim"`
	ScreenName          string           `json:"screenName" conform:"trim"`
	Subject             string           `json:"subject" conform:"trim"`
	Content             Content          `json:"content"`
	CommentCount        int              `json:"commentCount"`
	TagIDs              tags.TagIDs      `json:"tags"`
	UpVotes             int              `json:"upVotes"`
	DownVotes           int              `json:"downVotes"`
	PostCommentTrack    PostCommentTrack `json:"postCommentTrack,omitempty"`
	Comments            Comments         `json:"comments,omitempty"`
	NextCommentPage     *NextPage        `json:"nextCommentPage"`
	ReportedCount       int              `json:"reportedCount"`
	Obfuscated          bool             `json:"obfuscated"`
	Reviewed            bool             `json:"reviewed"`
	ReviewComment       string           `json:"reviewComment"`
	ReviewedDatetime    time.Time        `json:"reviewedDatetime,omitempty"`
	ReportedUsers       []ReportedUser   `json:"reportedUsers"`
	Deleted             bool             `json:"deleted,omitempty"`
	Video               PostVideo        `json:"video,omitempty"`
	IsAdminPost         bool             `json:"isAdminPost"`
}

type PostVideo struct {
	ReferenceId string `json:"referenceId,omitempty"`
	Source      string `json:"source"`
	Url         string `json:"url"`
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
		HiveID:              uint64(r.Number(math.MaxInt32)),
		PostID:              uint64(r.Number(math.MaxInt32)),
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

func PostVideoFromDB(p *dbmodels.PostVideo) PostVideo {
	out := PostVideo{
		ReferenceId: p.ReferenceID.String,
		Url:         p.URL,
		Source:      p.Source,
	}

	return out
}

func PostFromDB(p *dbmodels.Post) Post {
	out := Post{
		HiveID:              p.HiveID,
		IsPinnedPost:        p.Pinned,
		PostID:              p.PostID,
		PostDatetime:        p.CreatedAt,
		LastCommentDatetime: p.LastCommentTS,
		//Edits:               nil,
		ImpartWealthID: p.ImpartWealthID,
		//ScreenName:          p.sc,
		Subject:      p.Subject,
		Content:      Content{Markdown: p.Content},
		CommentCount: p.CommentCount,
		//TagIDs:              nil,
		UpVotes:   p.UpVoteCount,
		DownVotes: p.DownVoteCount,
		//PostCommentTrack:    PostCommentTrack{},
		//Comments:            nil,
		//NextCommentPage:     nil,
		ReportedCount: p.ReportedCount,
		Obfuscated:    p.Obfuscated,
		Reviewed:      p.Reviewed,
		ReviewComment: p.ReviewComment.String,
	}
	if p.R.ImpartWealth != nil {
		out.ScreenName = p.R.ImpartWealth.ScreenName
	}
	if p.ReviewedAt.Valid {
		out.ReviewedDatetime = p.ReviewedAt.Time
	}
	if len(p.R.Tags) > 0 {
		out.TagIDs = make([]int, len(p.R.Tags), len(p.R.Tags))
		for i, tId := range p.R.Tags {
			out.TagIDs[i] = int(tId.TagID)
		}
	}
	if len(p.R.PostReactions) > 0 {
		out.PostCommentTrack = PostCommentTrackFromDB(p.R.PostReactions[0], nil)
	}
	if len(p.R.Comments) > 0 {

	}
	if (p.DeletedAt != null.Time{}) {
		out.Deleted = true
	}
	if p.R.PostVideos != nil && len(p.R.PostVideos) > 0 {
		out.Video = PostVideoFromDB(p.R.PostVideos[0])
	}

	// check the user is blocked
	if p.R.ImpartWealth != nil && p.R.ImpartWealth.Blocked {
		out.ScreenName = "[deleted user]"
	}
	//check the user is admin
	if p.R.ImpartWealth != nil && p.R.ImpartWealth.Admin {
		out.IsAdminPost = true
	} else {
		out.IsAdminPost = false
	}
	return out
}

func PostsFromDB(dbPosts dbmodels.PostSlice) Posts {
	out := make(Posts, len(dbPosts), len(dbPosts))
	for i, p := range dbPosts {
		out[i] = PostFromDB(p)
	}
	return out
}

func (p Post) ToDBModel() *dbmodels.Post {
	out := &dbmodels.Post{
		PostID:         p.PostID,
		HiveID:         p.HiveID,
		ImpartWealthID: p.ImpartWealthID,
		Pinned:         p.IsPinnedPost,
		CreatedAt:      p.PostDatetime,
		Subject:        p.Subject,
		Content:        p.Content.Markdown,
		LastCommentTS:  p.LastCommentDatetime,
		CommentCount:   p.CommentCount,
		UpVoteCount:    p.UpVotes,
		DownVoteCount:  p.DownVotes,
	}

	return out
}

func PostCommentTrackFromPostReaction(r *dbmodels.PostReaction) PostCommentTrack {
	out := PostCommentTrack{
		ImpartWealthID: r.ImpartWealthID,
		ContentID:      r.PostID,
		PostID:         r.PostID,
		UpVoted:        r.Upvoted,
		DownVoted:      r.Downvoted,
		VotedDatetime:  r.UpdatedAt,
	}

	return out
}

type PostNotificationInput struct {
	Ctx        context.Context
	CommentID  uint64
	ActionType types.Type // Report,upvote,downvote, take vote
	ActionData string
	PostID     uint64
}

type PostNotificationBuildDataOutput struct {
	Alert             impart.Alert
	PostOwnerWealthID string
}

func PostsWithlimit(dbPosts dbmodels.PostSlice, limit int) Posts {
	out := make(Posts, limit, limit)
	for i, p := range dbPosts {
		if i >= limit {
			return out
		}
		out[i] = PostFromDB(p)
	}
	return out
}
