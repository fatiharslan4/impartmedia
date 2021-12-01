package models

import (
	"time"

	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
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
	SearchIDs []string
	SortBy    string
	SortOrder string
	Hive      int
}

type UserDetails []UserDetail
type UserDetail struct {
	ImpartWealthID   string      `json:"impartWealthId"`
	ScreenName       null.String `json:"screen_name"  `
	Email            null.String `json:"email" `
	CreatedAt        time.Time   `json:"created_at" `
	Admin            bool        `json:"admin" `
	Post             uint64      `json:"post" `
	HiveId           string      `json:"hive_id" `
	Household        string      `json:"household" `
	Dependents       string      `json:"dependents" `
	Generation       string      `json:"generation" `
	Gender           string      `json:"gender" `
	Race             string      `json:"race" `
	Financialgoals   string      `json:"financialgoals" `
	Industry         string      `json:"industry"`
	Career           string      `json:"career"`
	Income           string      `json:"income"`
	LastloginAt      string      `json:"lastlogin_at"`
	SuperAdmin       bool        `json:"super_admin"`
	AnswerIds        string      `json:"answer_ids"`
	EmploymentStatus string      `json:"employment_status"`
	// List             null.Uint64 `json:"list"`
	// Waitlist         bool        `json:"waitlist"`
}

type PagedUserResponse struct {
	UserDetails UserDetails `json:"users"`
	NextPage    *NextPage   `json:"nextPage"`
}

type PostDetails []PostDetail
type PostDetail struct {
	PostID         uint64    `json:"postid"`
	ScreenName     string    `json:"screen_name"  `
	Email          string    `json:"email" `
	PostDatetime   time.Time `json:"created_at" `
	HiveID         uint64    `json:"hive_id" `
	Pinned         bool      `json:"pinned" `
	Reported       bool      `json:"reported" `
	CommentCount   int       `json:"comment_count" `
	PostContent    string    `json:"content" `
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
	UrlPostUrl     string    `json:"url_post_url" `
	Tags           string    `json:"tag" `
}

type UserUpdate struct {
	Type   string     `json:"type,omitempty"`
	Action string     `json:"action"`
	HiveID uint64     `json:"hiveID,omitempty"`
	Users  []UserData `json:"users,omitempty"`
}

type UserData struct {
	ImpartWealthID string `json:"impartWealthId"`
	Status         bool   `json:"status"`
	Message        string `json:"message,omitempty"`
	Value          int    `json:"value"`
	ScreenName     string `json:"screen_name"  `
}

type PostUpdate struct {
	Action string     `json:"action"`
	Posts  []PostData `json:"posts,omitempty"`
}

type PostData struct {
	PostID  uint64 `json:"postID,omitempty"`
	Status  bool   `json:"status"`
	Message string `json:"message,omitempty"`
	Title   string `json:"title"`
}

type HiveUpdate struct {
	Action string     `json:"action"`
	Hives  []HiveData `json:"hives,omitempty"`
}

type HiveData struct {
	HiveID  uint64 `json:"hiveID,omitempty"`
	Status  bool   `json:"status"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name"`
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
	out.UrlPostUrl = "NA"
	out.Tags = "NA"
	if p.R.PostVideos != nil && len(p.R.PostVideos) > 0 {
		out.VideoType = p.R.PostVideos[0].Source
		if out.VideoType == "youtube" {
			out.VideoType = "YouTube"
		} else if out.VideoType == "vimeo" {
			out.VideoType = "Vimeo"
		}
		out.VideoUrl = p.R.PostVideos[0].URL
	}
	if p.R.PostUrls != nil && len(p.R.PostUrls) > 0 {
		out.UrlImage = p.R.PostUrls[0].ImageUrl
		out.UrlTitle = p.R.PostUrls[0].Title
		out.UrlDescription = p.R.PostUrls[0].Description
		out.UrlPostUrl = p.R.PostUrls[0].URL.String
	}
	if p.R.PostFiles != nil && len(p.R.PostFiles) > 0 {
		out.ImagePath = p.R.PostFiles[0].R.FidFile.URL
	}
	if len(p.R.Tags) > 0 {
		out.Tags = p.R.Tags[0].Name
	}
	return out
}

type PagedPostResponse struct {
	PostDetails PostDetails `json:"posts"`
	NextPage    *NextPage   `json:"nextPage"`
}

type PagedHiveResponse struct {
	Hive     HiveDetails `json:"hives"`
	NextPage *NextPage   `json:"nextPage"`
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

type PagedUserUpdateResponse struct {
	Users *UserUpdate `json:"users"`
}

type PagedPostUpdateResponse struct {
	Posts *PostUpdate `json:"posts"`
}

type PagedHiveUpdateResponse struct {
	Hives *HiveUpdate `json:"hives"`
}

type HiveDetails []HiveDetail
type HiveDetail struct {
	HiveId                               uint64      `json:"hive_id"`
	Name                                 string      `json:"name"  `
	CreatedAt                            null.Time   `json:"created_at" `
	Users                                null.Uint64 `json:"users" `
	HouseholdSingle                      null.Uint64 `json:"household_single" `
	HouseholdSingleroommates             null.Uint64 `json:"household_singleroommates" `
	HouseholdPartner                     null.Uint64 `json:"household_partner" `
	HouseholdMarried                     null.Uint64 `json:"household_married" `
	HouseholdSharedcustody               null.Uint64 `json:"household_sharedcustody" `
	DependentsNone                       null.Uint64 `json:"dependents_none" `
	DependentsPreschool                  null.Uint64 `json:"dependents_preschool" `
	DependentsSchoolage                  null.Uint64 `json:"dependents_schoolage" `
	DependentsPostschool                 null.Uint64 `json:"dependents_postschool" `
	DependentsParents                    null.Uint64 `json:"dependents_parents" `
	DependentsOther                      null.Uint64 `json:"dependents_other" `
	GenerationGenz                       null.Uint64 `json:"generation_genz" `
	GenerationMillennial                 null.Uint64 `json:"generation_millennial" `
	GenerationGenx                       null.Uint64 `json:"generation_genx" `
	GenerationBoomer                     null.Uint64 `json:"generation_boomer" `
	GenderWoman                          null.Uint64 `json:"gender_woman" `
	GenderMan                            null.Uint64 `json:"gender_man" `
	GenderNonbinary                      null.Uint64 `json:"gender_nonbinary" `
	GenderNotlisted                      null.Uint64 `json:"gender_notlisted" `
	RaceAmindianalnative                 null.Uint64 `json:"race_amindianalnative" `
	RaceAsianpacislander                 null.Uint64 `json:"race_asianpacislander" `
	RaceBlack                            null.Uint64 `json:"race_black" `
	RaceHispanic                         null.Uint64 `json:"race_hispanic" `
	RaceSwasiannafrican                  null.Uint64 `json:"race_swasiannafrican" `
	RaceWhite                            null.Uint64 `json:"race_white" `
	FinancialgoalsRetirement             null.Uint64 `json:"financialgoals_retirement" `
	FinancialgoalsSavecollege            null.Uint64 `json:"financialgoals_savecollege" `
	FinancialgoalsHouse                  null.Uint64 `json:"financialgoals_house" `
	FinancialgoalsPhilanthropy           null.Uint64 `json:"financialgoals_philanthropy" `
	FinancialgoalsGenerationalwealth     null.Uint64 `json:"financialgoals_generationalwealth" `
	IndustryAgriculture                  null.Uint64 `json:"fndustry_agriculture" `
	IndustryBusiness                     null.Uint64 `json:"industry_business" `
	IndustryConstruction                 null.Uint64 `json:"industry_construction" `
	IndustryEducation                    null.Uint64 `json:"industry_education" `
	IndustryEntertainmentgaming          null.Uint64 `json:"industry_entertainmentgaming" `
	IndustryFinancensurance              null.Uint64 `json:"industry_financensurance" `
	IndustryFoodhospitality              null.Uint64 `json:"industry_foodhospitality" `
	IndustryGovernmentpublicservices     null.Uint64 `json:"industry_governmentpublicservices" `
	IndustryHealthservices               null.Uint64 `json:"industry_healthservices" `
	IndustryLegal                        null.Uint64 `json:"industry_legal" `
	IndustryNaturalresources             null.Uint64 `json:"industry_naturalresources" `
	IndustryPersonalprofessionalservices null.Uint64 `json:"industry_personalprofessionalservices" `
	IndustryRealestatehousing            null.Uint64 `json:"industry_realestatehousing" `
	IndustryRetailecommerce              null.Uint64 `json:"industry_retailecommerce" `
	IndustrySafetysecurity               null.Uint64 `json:"industry_safetysecurity" `
	IndustryTransportation               null.Uint64 `json:"industry_transportation" `
	CareerEntrylevel                     null.Uint64 `json:"career_entrylevel" `
	CareerMidlevel                       null.Uint64 `json:"career_midlevel" `
	CareerManagement                     null.Uint64 `json:"career_management" `
	CareerUppermanagement                null.Uint64 `json:"career_uppermanagement" `
	CareerBusinessowner                  null.Uint64 `json:"career_businessowner" `
	CareerOther                          null.Uint64 `json:"career_other" `
	IncomeIncome0                        null.Uint64 `json:"income_income0" `
	IncomeIncome1                        null.Uint64 `json:"income_income1" `
	IncomeIncome2                        null.Uint64 `json:"income_income2" `
	IncomeIncome3                        null.Uint64 `json:"income_income3" `
	IncomeIncome4                        null.Uint64 `json:"income_income4" `
	IncomeIncome5                        null.Uint64 `json:"income_income5" `
	EmploymentstatusFulltime             null.Uint64 `json:"employmentstatus_fulltime" `
	EmploymentstatusParttime             null.Uint64 `json:"employmentstatus_parttime" `
	EmploymentstatusUnemployed           null.Uint64 `json:"employmentstatus_unemployed" `
	EmploymentstatusSelf                 null.Uint64 `json:"employmentstatus_self" `
	EmploymentstatusHomemaker            null.Uint64 `json:"employmentstatus_homemaker" `
	EmploymentstatusStudent              null.Uint64 `json:"employmentstatus_student" `
	EmploymentstatusRetired              null.Uint64 `json:"employmentstatus_retired" `
	IncomeIncome6                        null.Uint64 `json:"income_income6" `
	IncomeIncome7                        null.Uint64 `json:"income_income7" `
	IncomeIncome8                        null.Uint64 `json:"income_income8" `
}
