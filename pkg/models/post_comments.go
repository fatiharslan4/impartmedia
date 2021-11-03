package models

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/volatiletech/null/v8"
)

type PostComments []PostComment
type PostComment struct {
	Type                string           `json:"type"`
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
	ReportedCount       int              `json:"reportedCount"`
	Obfuscated          bool             `json:"obfuscated"`
	Reviewed            bool             `json:"reviewed"`
	ReviewComment       string           `json:"reviewComment"`
	ReviewedDatetime    time.Time        `json:"reviewedDatetime,omitempty"`
	ReportedUsers       []ReportedUser   `json:"reportedUsers"`
	Deleted             bool             `json:"deleted,omitempty"`
	Video               PostVideo        `json:"video,omitempty"`
	IsAdminPost         bool             `json:"isAdminPost"`
	CommentID           uint64           `json:"commentId,omitempty"`
	CommentDatetime     time.Time        `json:"commentDatetime,omitempty"`
	ParentCommentID     uint64           `json:"parentCommentId"`
	UpdatedDate         time.Time        `json:"updatedDate"`
	Files               []File           `json:"file,omitempty"`
	Url                 string           `json:"url,omitempty"`
	UrlData             PostUrl          `json:"urlData,omitempty"`
	FirstName           string           `json:"firstName,omitempty"`
	LastName            string           `json:"lastName,omitempty"`
	FullName            string           `json:"fullName,omitempty"`
	AvatarBackground    string           `json:"avatarBackground,omitempty"`
	AvatarLetter        string           `json:"avatarLetter,omitempty"`
	Admin               bool             `json:"admin,omitempty"`
}

func PostCommentsLimit(dbPosts dbmodels.PostSlice, dbcomments dbmodels.CommentSlice, limit int) PostComments {
	postout := make(PostComments, len(dbPosts))
	cmntout := make(PostComments, len(dbcomments))
	for i, p := range dbPosts {
		postout[i] = PostCommentPostFromDB(p, nil)
	}
	for i, cmnt := range dbcomments {
		cmntout[i] = PostCommentPostFromDB(nil, cmnt)
	}
	out := append(postout, cmntout...)
	return out
}

func PostCommentPostFromDB(p *dbmodels.Post, c *dbmodels.Comment) PostComment {
	// var out PostComment
	if p != nil {
		out := PostComment{
			Type:                "Post",
			HiveID:              p.HiveID,
			IsPinnedPost:        p.Pinned,
			PostID:              p.PostID,
			PostDatetime:        p.CreatedAt,
			LastCommentDatetime: p.LastCommentTS,
			ImpartWealthID:      p.ImpartWealthID,
			Subject:             p.Subject,
			Content:             Content{Markdown: p.Content},
			CommentCount:        p.CommentCount,
			UpVotes:             p.UpVoteCount,
			DownVotes:           p.DownVoteCount,
			ReportedCount:       p.ReportedCount,
			Obfuscated:          p.Obfuscated,
			Reviewed:            p.Reviewed,
			ReviewComment:       p.ReviewComment.String,
		}
		if p.R.ImpartWealth != nil {
			out.ScreenName = p.R.ImpartWealth.ScreenName
			out.FirstName = strings.Title(p.R.ImpartWealth.FirstName)
			out.LastName = strings.Title(p.R.ImpartWealth.LastName)
			out.FullName = strings.Title(fmt.Sprintf("%s %s", p.R.ImpartWealth.FirstName, p.R.ImpartWealth.LastName))
			out.AvatarBackground = p.R.ImpartWealth.AvatarBackground
			out.AvatarLetter = p.R.ImpartWealth.AvatarLetter
			out.Admin = p.R.ImpartWealth.Admin
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
			out.UpdatedDate = p.R.PostReactions[0].CreatedAt
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
			out.ScreenName = string(types.AccountRemoved)
		}
		if p.R.ImpartWealth == nil {
			out.ScreenName = types.AccountDeleted.ToString()
		}
		//check the user is admin
		if p.R.ImpartWealth != nil && p.R.ImpartWealth.Admin {
			out.IsAdminPost = true
		} else {
			out.IsAdminPost = false
		}

		// post files
		if p.R.PostFiles != nil {
			out.Files = make([]File, 0)
			for _, f := range p.R.PostFiles {
				if f.R.FidFile != nil {
					out.Files = append(out.Files, PostFileToFile(f))
				}
			}
		}

		if p.R.PostUrls != nil && len(p.R.PostUrls) > 0 {
			out.UrlData = PostUrlFromDB(p.R.PostUrls[0])
		}

		return out

	}
	if c != nil {
		out := PostComment{
			Type:            "Comment",
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
			out.UpdatedDate = c.R.CommentReactions[0].CreatedAt
		}

		if (c.DeletedAt != null.Time{}) {
			out.Deleted = true
		}

		// check the user is blocked
		if c.R.ImpartWealth != nil && c.R.ImpartWealth.Blocked {
			out.ScreenName = string(types.AccountRemoved)
		}
		if c.R.ImpartWealth == nil {
			out.ScreenName = types.AccountDeleted.ToString()
		}

		return out
	}
	return PostComment{}
}

func (postcomments PostComments) SortDescending() {
	sort.Slice(postcomments, func(i, j int) bool {
		return postcomments[i].UpdatedDate.After(postcomments[j].UpdatedDate)
	})
}

func PostsCommentsWithLimit(postcomments PostComments, limit int) PostComments {
	out := make(PostComments, limit, limit)
	for i, p := range postcomments {
		if i >= limit {
			return out
		}
		out[i] = p
	}
	return out
}

func CountPostComnt(posts PostComments) (int, int) {
	pstCnt := 0
	cmntCnt := 0
	for _, post := range posts {
		if post.Type == "Post" {
			pstCnt += 1
		}
		if post.Type == "Comment" {
			cmntCnt += 1
		}
	}
	return pstCnt, cmntCnt
}
