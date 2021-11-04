package models

import (
	"encoding/json"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/leebenson/conform"
	"github.com/segmentio/ksuid"
)

type NextPage struct {
	Offset        int `json:"offset"`
	OffsetPost    int `json:"offsetPost"`
	OffsetComment int `json:"offsetComment"`
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
	HiveID uint64 `json:"hiveId,omitempty"`
	// HiveID          uint64 `json:"hiveId" jsonschema:"minLength=27,maxLength=27"`
	HiveName        string `json:"hiveName" conform:"trim" jsonschema:"minLength=3,maxLength=60"`
	HiveDescription string `json:"hiveDescription" conform:"trim,omitempty" jsonschema:"minLength=0,maxLength=5000"`
	//Administrators    []HiveAdmin       `json:"administrators"`
	HiveDistributions HiveDistributions `json:"hiveDistributions,omitempty"`
	//Metrics                        HiveMetrics         `json:"metrics,omitempty"`
	PinnedPostID   uint64              `json:"pinnedPostId,omitempty"`
	TagComparisons tags.TagComparisons `json:"tagComparisons,omitempty"`
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
	Reported       bool      `json:"reported"`
	ReportedReason string    `json:"reportedReason" conform:"trim"`
}

func PostCommentTrackFromDB(p *dbmodels.PostReaction, c *dbmodels.CommentReaction) PostCommentTrack {
	var out PostCommentTrack
	if p == nil {
		out = PostCommentTrack{
			ImpartWealthID: c.ImpartWealthID,
			ContentID:      c.CommentID,
			PostID:         c.PostID,
			UpVoted:        c.Upvoted,
			DownVoted:      c.Downvoted,
			VotedDatetime:  c.UpdatedAt,
			Reported:       c.Reported,
		}
		if c.ReportedReason.Valid {
			out.ReportedReason = c.ReportedReason.String
		}
	} else {
		//is a post
		out = PostCommentTrack{
			ImpartWealthID: p.ImpartWealthID,
			ContentID:      p.PostID,
			PostID:         p.PostID,
			UpVoted:        p.Upvoted,
			DownVoted:      p.Downvoted,
			VotedDatetime:  p.UpdatedAt,
			Reported:       p.Reported,
		}
		if p.ReportedReason.Valid {
			out.ReportedReason = p.ReportedReason.String
		}
	}

	return out
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
	Markdown string `json:"markdown" jsonschema:"maxLength=300000" conform:"trim"`
}

type Edits []Edit
type Edit struct {
	ImpartWealthID string    `json:"impartWealthId" jsonschema:"minLength=27,maxLength=27" conform:"trim"`
	ScreenName     string    `json:"screenName" conform:"trim"`
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
		// HiveID:       h.HiveID,
		Name:         h.HiveName,
		Description:  h.HiveDescription,
		PinnedPostID: null.Uint64From(h.PinnedPostID),
		//TagComparisons:       null.JSON{},
		//HiveDistributions:    null.JSON{},
		CreatedAt: null.TimeFrom(impart.CurrentUTC()),
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

type HiveRules []HiveRule
type HiveRule struct {
	RuleID    uint64         `json:"ruleId,omitempty"`
	RuleName  string         `json:"ruleName" conform:"trim,omitempty,ucfirst" jsonschema:"minLength=3,maxLength=60"`
	Status    bool           `json:"status,omitempty"`
	Limit     int64          `json:"limit,omitempty"`
	UserCount int64          `json:"userCount,omitempty"`
	Question  []Question     `json:"questions,omitempty"`
	Criteria  []CriteriaData `json:"criteria,omitempty"`
	Hive      []Hive         `json:"hive,omitempty"`
	HiveID    null.Uint64    `json:"hiveId,omitempty"`
}

type CriteriaData struct {
	AnswerID []uint `json:"answerID,omitempty"`
}

func (hiverule HiveRule) ToDBModel() (*dbmodels.HiveRule, error) {
	rule := &dbmodels.HiveRule{
		Name:     hiverule.RuleName,
		Status:   hiverule.Status,
		MaxLimit: int64(hiverule.Limit),
		HiveID:   hiverule.HiveID,
	}
	return rule, nil
}

func HiveRulesFromDB(dbHiveRules dbmodels.HiveRuleSlice) (HiveRules, error) {
	out := make(HiveRules, len(dbHiveRules), len(dbHiveRules))
	for pos, dbh := range dbHiveRules {
		hive, err := HiveRuleFromDB(dbh)
		if err != nil {
			return out, err
		}
		out[pos] = *hive
	}
	return out, nil
}

func HiveRuleFromDB(dbHive *dbmodels.HiveRule) (*HiveRule, error) {
	out := HiveRule{
		RuleID:    dbHive.RuleID,
		RuleName:  dbHive.Name,
		UserCount: dbHive.NoOfUsers,
		Limit:     dbHive.MaxLimit,
	}
	if dbHive.R.Hives != nil {
		hives := make(Hives, len(dbHive.R.Hives))
		for pos, hive := range dbHive.R.Hives {
			outhive := Hive{
				HiveID:   hive.HiveID,
				HiveName: hive.Name,
			}
			hives[pos] = outhive
		}
		out.Hive = hives
	}
	if dbHive.R.RuleHiveRulesCriteria != nil {
		var questions []Question
		var answer []Answer
		var questionId uint
		outhiveQuestion := Question{}

		for _, hive := range dbHive.R.RuleHiveRulesCriteria {
			if questionId == hive.QuestionID {
				continue
			}
			answer = nil
			outhiveQuestion = Question{
				Id:           hive.QuestionID,
				Name:         hive.R.Question.QuestionName,
				QuestionText: hive.R.Question.Text,
			}
			for _, ansr := range dbHive.R.RuleHiveRulesCriteria {
				if ansr.QuestionID == hive.QuestionID {
					ans := Answer{
						Id:   ansr.R.Answer.AnswerID,
						Name: ansr.R.Answer.AnswerName,
					}
					answer = append(answer, ans)
					outhiveQuestion.Answers = answer
				}
			}
			questions = append(questions, outhiveQuestion)
			questionId = hive.QuestionID
		}
		out.Question = questions
	}

	return &out, nil
}

type PagedHiveRoleResponse struct {
	HiveRules HiveRuleLists `json:"hiveRules"`
	NextPage  *NextPage     `json:"nextPage"`
}

type GetHiveInput struct {
	// Limit is the maximum number of records that should be returns.  The API can optionally return
	// less than Limit, if DynamoDB decides the items read were too large.
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

func HiveRuleDBToModel(hiverule *dbmodels.HiveRule) (*HiveRule, error) {
	rule := &HiveRule{
		RuleID:   hiverule.RuleID,
		RuleName: hiverule.Name,
		Status:   hiverule.Status,
		Limit:    hiverule.MaxLimit,
		HiveID:   hiverule.HiveID,
	}
	return rule, nil
}

type HiveRuleLists []HiveRuleList
type HiveRuleList struct {
	RuleId           uint64 `json:"rule_id"`
	Name             string `json:"name"`
	MaxLimit         int    `json:"max_limit" `
	NoOfUsers        int    `json:"no_of_users" `
	Status           bool   `json:"status" `
	HiveId           string `json:"hive_id" `
	HiveName         string `json:"hive_name" `
	Household        string `json:"household" `
	Dependents       string `json:"dependents" `
	Generation       string `json:"generation" `
	Gender           string `json:"gender" `
	Race             string `json:"race" `
	Financialgoals   string `json:"financialgoals" `
	Industry         string `json:"industry"`
	Career           string `json:"career"`
	Income           string `json:"income"`
	EmploymentStatus string `json:"employment_status"`
}
