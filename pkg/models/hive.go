package models

import (
	"encoding/json"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/leebenson/conform"
	"github.com/segmentio/ksuid"
)

type NextPage struct {
	Offset int `json:"offset"`
}

// HiveMemberships are primarily an attribute of the profile for a customer
// Said another way - Hives don't have members, members have hives with which they belong.
type HiveMemberships []HiveMembership
type HiveMembership struct {
	HiveID   uint64 `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	HiveName string `json:"hiveName,omitempty" conform:"ucfirst,trim"`
}

//type HiveAdmin struct {
//	ImpartWealthID string    `json:"impartWealthId"`
//	ScreenName     string    `json:"screenName" conform:"trim"`
//	AdminDatetime  time.Time `json:"adminDatetime,omitempty"`
//}

type Hives []Hive

// Hive represents the top level organization of a hive community
type Hive struct {
	HiveID          uint64 `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	HiveName        string `json:"hiveName" conform:"ucfirst,trim"`
	HiveDescription string `json:"hiveDescription" conform:"trim"`
	//Administrators    []HiveAdmin       `json:"administrators"`
	HiveDistributions HiveDistributions `json:"hiveDistributions,omitempty"`
	//Metrics                        HiveMetrics         `json:"metrics,omitempty"`
	PinnedPostID   uint64              `json:"pinnedPostId"`
	TagComparisons tags.TagComparisons `json:"tagComparisons"`
	//PinnedPostNotificationTopicARN string              `json:"pinnedPostNotificationTopicARN"`
}

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
	ContentID      uint64    `json:"contentId" jsonschema:"minLength=27,maxLength=27"`
	PostID         uint64    `json:"postId,omitempty"`
	UpVoted        bool      `json:"upVoted"`
	DownVoted      bool      `json:"downVoted"`
	VotedDatetime  time.Time `json:"votedDatetime,omitempty"`
}

func PostCommentTrackFromDB(p *dbmodels.PostReaction, c *dbmodels.CommentReaction) PostCommentTrack {
	if p == nil {
		return PostCommentTrack{
			ImpartWealthID: c.ImpartWealthID,
			ContentID:      c.CommentID,
			PostID:         c.PostID,
			UpVoted:        c.Upvoted,
			DownVoted:      c.Downvoted,
			VotedDatetime:  c.UpdatedTS,
		}
	}
	//is a post
	return PostCommentTrack{
		ImpartWealthID: p.ImpartWealthID,
		ContentID:      p.PostID,
		PostID:         p.PostID,
		UpVoted:        p.Upvoted,
		DownVoted:      p.Downvoted,
		VotedDatetime:  p.UpdatedTS,
	}

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

func (h Hive) ToJson() string {
	b, _ := json.MarshalIndent(&h, "", "\t")
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
		HiveID:          uint64(r.Number(math.MaxInt32)),
		HiveName:        r.Noun(),
		HiveDescription: r.Noun() + r.SillyName(),
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

func HivesFromDB(dbHives dbmodels.HiveSlice) (Hives, error) {
	out := make(Hives, len(dbHives), len(dbHives))
	for i, dbh := range dbHives {
		h, err := HiveFromDB(dbh)
		if err != nil {
			return out, err
		}
		out[i] = h
	}
	return out, nil
}

func HiveFromDB(dbHive *dbmodels.Hive) (Hive, error) {
	out := Hive{
		HiveID:          dbHive.HiveID,
		HiveName:        dbHive.Name,
		HiveDescription: dbHive.Description,
		PinnedPostID:    dbHive.PinnedPostID.Uint64,
	}

	if err := dbHive.HiveDistributions.Unmarshal(&out.HiveDistributions); err != nil {
		return out, err
	}

	if err := dbHive.TagComparisons.Unmarshal(&out.TagComparisons); err != nil {
		return out, err
	}

	return out, nil
}

func (h Hive) ToDBModel() (*dbmodels.Hive, error) {
	dbh := &dbmodels.Hive{
		HiveID:       h.HiveID,
		Name:         h.HiveName,
		Description:  h.HiveDescription,
		PinnedPostID: null.Uint64From(h.PinnedPostID),
		//TagComparisons:       null.JSON{},
		//HiveDistributions:    null.JSON{},
	}
	err := dbh.HiveDistributions.Marshal(&h.HiveDistributions)
	if err != nil {
		return nil, err
	}
	err = dbh.TagComparisons.Marshal(&h.TagComparisons)
	if err != nil {
		return nil, err
	}
	return dbh, nil
}
