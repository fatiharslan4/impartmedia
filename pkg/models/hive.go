package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	r "github.com/Pallinder/go-randomdata"
	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/leebenson/conform"
	"github.com/segmentio/ksuid"
)

const OffsetContentIDQueryParam = "offsetContentId"
const OffsetTimestampQueryParam = "offsetTimestamp"

type NextPage struct {
	ContentID string    `json:"offsetContentId"`
	Timestamp time.Time `json:"offsetTimestamp"`
}

func NextPageFromContext(ctx *gin.Context) (*NextPage, error) {
	offsetContentId := ctx.Query(OffsetContentIDQueryParam)
	offsetTimestamp := ctx.Query(OffsetTimestampQueryParam)

	contentIdPresent := len(strings.TrimSpace(offsetContentId)) > 0
	timestampPresent := len(strings.TrimSpace(offsetTimestamp)) > 0

	if (!contentIdPresent) && !timestampPresent {
		return nil, nil
	}

	if (contentIdPresent && !timestampPresent) || (!contentIdPresent && timestampPresent) {
		return nil, fmt.Errorf("invalid offset - both %s and %s must be provided", OffsetContentIDQueryParam, OffsetTimestampQueryParam)
	}

	parsedTimestamp, err := time.Parse(time.RFC3339, offsetTimestamp)
	if err != nil {
		return nil, err
	}

	return &NextPage{
		ContentID: offsetContentId,
		Timestamp: parsedTimestamp,
	}, nil
}

// HiveMemberships are primarily an attribute of the profile for a customer
// Said another way - Hives don't have members, members have hives with which they belong.
type HiveMemberships []HiveMembership
type HiveMembership struct {
	HiveID       string    `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	HiveName     string    `json:"hiveName,omitempty" conform:"ucfirst,trim"`
	JoinDatetime time.Time `json:"joinDateTime,omitempty"`
}

type HiveAdmin struct {
	ImpartWealthID string    `json:"impartWealthId"`
	ScreenName     string    `json:"screenName" conform:"trim"`
	AdminDatetime  time.Time `json:"adminDatetime,omitempty"`
}

type Hives []Hive

// Hive represents the top level organization of a hive community
type Hive struct {
	HiveID                         string              `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	HiveName                       string              `json:"hiveName" conform:"ucfirst,trim"`
	HiveDescription                string              `json:"hiveDescription" conform:"trim"`
	Administrators                 []HiveAdmin         `json:"administrators"`
	HiveDistributions              HiveDistributions   `json:"hiveDistributions,omitempty"`
	Metrics                        HiveMetrics         `json:"metrics,omitempty"`
	PinnedPostID                   string              `json:"pinnedPostId"`
	TagComparisons                 tags.TagComparisons `json:"tagComparisons"`
	PinnedPostNotificationTopicARN string              `json:"pinnedPostNotificationTopicARN"`
}

//type HiveDetails struct {
//	AgeRange               string `json:"ageRange"`
//	GenderDistribution     string `json:"genderDistribution"`
//	MaritalDistribution    string `json:"maritalDistribution"`
//	EmploymentDistribution string `json:"employmentDistribution"`
//	Income                 string `json:"income"`
//	Location               string `json:"location"`
//	HasChildren            string `json:"hasChildren"`
//	HasEmergencyFunds      string `json:"hasEmergencyFunds"`
//	HasSavings             string `json:"hasSavings"`
//	HasWill                string `json:"hasWill"`
//}

type HiveMetrics struct {
	MemberCount int `json:"memberCount"`
	PostCount   int `json:"postCount"`
}

type HiveDistributions []HiveDistribution
type HiveDistribution struct {
	DisplayText  string `json:"displayText"`
	DisplayValue string `json:"displayValue"`
	SortValue    int    `json:"sortValue"`
}

func (hds HiveDistributions) Sort() {
	sort.Slice(hds, func(i, j int) bool {
		return hds[i].SortValue < hds[j].SortValue
	})
}

func (hds HiveDistributions) IsSorted() bool {
	return sort.SliceIsSorted(hds, func(i, j int) bool {
		return hds[i].SortValue < hds[j].SortValue
	})
}

func (hds *HiveDistributions) CleanEmptyValues() {
	for i, d := range *hds {
		if strings.TrimSpace(d.DisplayValue) == "" {
			hds.Pop(i)
		}
	}
}

func (hds *HiveDistributions) Pop(position int) {
	copy((*hds)[position:], (*hds)[position+1:])
	*hds = (*hds)[:len(*hds)-1]

}

// PostCommentTrack is used to track interaction with content; prevents multiple upvotes from occurring, etc.
// The ContentID is a ksuid, which is either a postID or a commentID - it is guaranteed unique.
// If the ContentID is a PostID, then PostID will be empty, while if contentID is a commentID, then PostID will be non-empty.
type PostCommentTrack struct {
	ImpartWealthID string    `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27"`
	ContentID      string    `json:"contentId" jsonschema:"minLength=27,maxLength=27"`
	HiveID         string    `json:"hiveId"`
	PostID         string    `json:"postId,omitempty"`
	UpVoted        bool      `json:"upVoted"`
	DownVoted      bool      `json:"downVoted"`
	VotedDatetime  time.Time `json:"votedDatetime,omitempty"`
	Saved          bool      `json:"saved,omitempty"`
}

// Latest returns the
func (e Edits) Latest() time.Time {
	var t = time.Unix(0, 0)
	for _, edit := range e {
		if edit.Datetime.After(t) {
			t = edit.Datetime
		}
	}
	return t
}

func (e Edits) SortAscending() {
	sort.Slice(e, func(i, j int) bool {
		return e[i].Datetime.Before(e[j].Datetime)
	})
}

func (e Edits) SortDescending() {
	sort.Slice(e, func(i, j int) bool {
		return e[i].Datetime.After(e[j].Datetime)
	})
}

func (p Hive) ToJson() string {
	b, _ := json.MarshalIndent(&p, "", "\t")
	return string(b)
}

func (h Hive) Equals(hc Hive) bool {
	return reflect.DeepEqual(h, hc)
}

func (h Hive) Copy() Hive {
	return h
}

func RandomHive() Hive {
	h := Hive{
		HiveID:          ksuid.New().String(),
		HiveName:        r.Noun(),
		HiveDescription: r.Noun() + r.SillyName(),
		Administrators: []HiveAdmin{
			{
				ImpartWealthID: ksuid.New().String(),
				ScreenName:     r.SillyName(),
				AdminDatetime:  impart.CurrentUTC(),
			},
		},
		//Details: HiveDetails{
		//	AgeRange:           "Generation X (1961 - 1981)",
		//	GenderDistribution: "whatever",
		//},
		Metrics: HiveMetrics{
			MemberCount: r.Number(1, 100),
			PostCount:   r.Number(100, 1000),
		},
	}
	conform.Strings(&h)
	return h
}

type Content struct {
	Markdown string `json:"markdown" jsonschema:"maxLength=300000"`
}

type Edits []Edit
type Edit struct {
	ImpartWealthID string    `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27"`
	ScreenName     string    `json:"screenName"`
	Datetime       time.Time `json:"datetime"`
	Notes          string    `json:"notes,omitempty" conform:"trim"`
	Deleted        bool      `json:"deleted"`
}

func RandomEdit() Edit {
	return Edit{
		ImpartWealthID: ksuid.New().String(),
		ScreenName:     r.SillyName(),
		Datetime:       impart.CurrentUTC(),
		Notes:          r.Paragraph(),
		Deleted:        false,
	}
}

func RandomContent(length int) Content {
	c := strings.Builder{}
	for i := 0; i < length; i++ {
		c.WriteString(r.Paragraph())
	}
	return Content{
		Markdown: c.String(),
	}
}
