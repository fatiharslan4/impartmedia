package models

import (
	"time"

	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
)

// GetAdminInputs is the input necessary
type GetAdminInputs struct {
	// Limit is the maximum number of records that should be returns.  The API can optionally return
	// less than Limit, if DynamoDB decides the items read were too large.
	Limit  int
	Offset int
	// search is the optional to filter on
	SearchKey string
}

type UserDetails []UserDetail
type UserDetail struct {
	ImpartWealthID string    `json:"impartWealthId"`
	ScreenName     string    `json:"screenName"  `
	Email          string    `json:"email" `
	CreatedAt      time.Time `json:"created_at" `
	Admin          bool      `json:"admin" `
	Post           uint64    `json:"post" `
	Hive           string    `json:"hive" `
	Household      string    `json:"household" `
	Dependents     string    `json:"dependents" `
	Generation     string    `json:"generation" `
	Gender         string    `json:"gender" `
	Race           string    `json:"race" `
	Financialgoals string    `json:"financialgoals" `
	LastLoginAt    null.Time `json:"last_login_at"`
}

type PagedUserResponse struct {
	UserDetails UserDetails `json:"users"`
	NextPage    *NextPage   `json:"nextPage"`
}

type PostDetails []PostDetail
type PostDetail struct {
	PostID         uint64    `json:"postid"`
	ScreenName     string    `json:"screenName"  `
	Email          string    `json:"email" `
	PostDatetime   time.Time `json:"created_at" `
	HiveID         uint64    `json:"hiveid" `
	Pinned         bool      `json:"pinned" `
	Reported       bool      `json:"reported" `
	CommentCount   int       `json:"commentcount" `
	PostContent    string    `json:"postcontent" `
	ImpartWealthID string    `json:"impartWealthId"`
	Subject        string    `json:"subject" `
	IsAdminPost    bool      `json:"adminpost" `
	Reviewed       bool      `json:"reviewed"`
}

func PostsData(dbPosts dbmodels.PostSlice) PostDetails {
	out := make(PostDetails, len(dbPosts), len(dbPosts))
	for i, p := range dbPosts {
		out[i] = PostsDataFromDB(p)
	}
	return out
}

func PostsDataFromDB(p *dbmodels.Post) PostDetail {
	out := PostDetail{
		HiveID:         p.HiveID,
		Pinned:         p.Pinned,
		PostID:         p.PostID,
		PostDatetime:   p.CreatedAt,
		ImpartWealthID: p.ImpartWealthID,
		Subject:        p.Subject,
		PostContent:    p.Content,
		CommentCount:   p.CommentCount,
		Reviewed:       p.Reviewed,
	}
	if p.ReportedCount > 0 {
		out.Reported = true
	} else {
		out.Reported = false
	}
	if p.R.ImpartWealth != nil {
		out.ScreenName = p.R.ImpartWealth.ScreenName
		out.Email = p.R.ImpartWealth.Email
	}

	// check the user is blocked
	if p.R.ImpartWealth != nil && p.R.ImpartWealth.Blocked {
		out.ScreenName = types.AccountRemoved.ToString()
		out.Email = types.AccountRemoved.ToString()
	}
	if p.R.ImpartWealth == nil {
		out.ScreenName = types.AccountDeleted.ToString()
		out.Email = types.AccountDeleted.ToString()
	}
	//check the user is admin
	if p.R.ImpartWealth != nil && p.R.ImpartWealth.Admin {
		out.IsAdminPost = true
	} else {
		out.IsAdminPost = false
	}
	return out
}

type PagedPostResponse struct {
	PostDetails PostDetails `json:"posts"`
	NextPage    *NextPage   `json:"nextPage"`
}
