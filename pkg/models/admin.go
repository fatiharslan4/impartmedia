package models

import (
	"time"

	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
)

// GetAdminInputs is the input necessary
type GetAdminInputs struct {
	// Limit is the maximum number of records that should be returns.  The API can optionally return
	// less than Limit, if DynamoDB decides the items read were too large.
	Limit  int
	Offset int
	// search is the optional to filter on
	SearchKey string
	SearchIDs string
	SortBy    string
	SortOrder string
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
	LastLoginAt    string    `json:"last_login_at"`
	SuperAdmin     bool      `json:"super_admin"`
	AnswerIds      string    `json:"answer_ids"`
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
	ImagePath      string    `json:"image_path" `
	VideoType      string    `json:"video_type" `
	VideoUrl       string    `json:"video_url" `
	UrlTitle       string    `json:"url_title" `
	UrlImage       string    `json:"url_image" `
	UrlDescription string    `json:"url_description" `
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
	}
	if p.Reviewed {
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
	out.VideoType = "NA"
	out.VideoUrl = "NA"
	out.UrlImage = "NA"
	out.UrlTitle = "NA"
	out.UrlDescription = "NA"
	out.ImagePath = "NA"
	if p.R.PostVideos != nil && len(p.R.PostVideos) > 0 {
		out.VideoType = p.R.PostVideos[0].Source
		out.VideoUrl = p.R.PostVideos[0].URL
	}
	if p.R.PostUrls != nil && len(p.R.PostUrls) > 0 {
		out.UrlImage = p.R.PostUrls[0].ImageUrl
		out.UrlTitle = p.R.PostUrls[0].Title
		out.UrlDescription = p.R.PostUrls[0].Description
	}
	if p.R.PostFiles != nil && len(p.R.PostFiles) > 0 {
		out.ImagePath = p.R.PostFiles[0].R.FidFile.URL
	}
	return out
}

type PagedPostResponse struct {
	PostDetails PostDetails `json:"posts"`
	NextPage    *NextPage   `json:"nextPage"`
}

type PagedHiveResponse struct {
	Hive     []map[string]interface{} `json:"hives"`
	NextPage *NextPage                `json:"nextPage"`
}

type MemberHives []MemberHive
type MemberHive struct {
	ImpartWealthID string `json:"impartWealthId"`
	MemberHiveId   uint64 `json:"member_hive_id"  `
}

type DemographicHivesCounts []DemographicHivesCount
type DemographicHivesCount struct {
	Count        int    `json:"count"`
	MemberHiveId uint64 `json:"member_hive_id"  `
}

type PagedFilterResponse struct {
	Filter impart.FilterEnum `json:"filter"`
}
