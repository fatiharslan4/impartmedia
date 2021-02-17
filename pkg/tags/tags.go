package tags

import (
	"encoding/json"
	"io/ioutil"
	"sort"

	"github.com/pkg/errors"
)

// Don't re-order these - these are effectively an ordered enum list - it will
// change behavior of existing tags.  this list is append only.
const (
	Unknown = iota
	IncomeID
	EmergencySavingsID
	EducationSavingsID
	RetirementSavingsID
	LifeInsuranceCoverageID
	OtherID
	NetWorthID
)

var ErrInvalidTagID = errors.New("unknown Tag ID")

type Tag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	LongName    string `json:"longName"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

type Tags []Tag
type TagIDs []int

func (t Tags) SortAscending() {
	sort.Slice(t, func(i, j int) bool {
		return t[i].SortOrder < t[j].SortOrder
	})
}

func AvailableTags() Tags {
	t := Tags{
		Income(),
		EmergencySavings(),
		EducationSavings(),
		RetirementSavings(),
		LifeInsuranceCoverage(),
		NetWorth(),
		Other(),
	}
	t.SortAscending()
	return t
}

type TagComparisons []TagComparison

func (t TagComparisons) SortAscending() {
	sort.Slice(t, func(i, j int) bool {
		return t[i].SortOrder < t[j].SortOrder
	})
}

type TagComparison struct {
	TagID              int         `json:"tagId"`
	SortOrder          int         `json:"sortOrder"`
	Percentile         int         `json:"percentile"`
	DisplayPercentile  string      `json:"displayPercentile"`
	Value              int         `json:"value"`
	DisplayValue       string      `json:"displayValue,omitempty"`
	DisplayScope       string      `json:"displayScope"`
	DisplayDescription string      `json:"displayDescription,omitempty"`
	Percentiles        Percentiles `json:"percentiles"`
}

type Percentiles []Percentile
type Percentile struct {
	Percent      int    `json:"percent"`
	HighValue    int    `json:"highValue"`
	DisplayValue string `json:"displayValue,omitempty"`
	Fill         bool   `json:"fill,omitempty"`
}

func (t Percentiles) SortAscending() {
	sort.Slice(t, func(i, j int) bool {
		return t[i].Percent < t[j].Percent
	})
}

func ImportComparisons(file string) (TagComparisons, error) {
	var out TagComparisons
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return out, err
	}

	err = json.Unmarshal(b, &out)
	return out, err
}

func FromID(tagID int) (Tag, error) {
	switch tagID {
	case IncomeID:
		return Income(), nil
	case EmergencySavingsID:
		return EmergencySavings(), nil
	case EducationSavingsID:
		return EducationSavings(), nil
	case RetirementSavingsID:
		return RetirementSavings(), nil
	case LifeInsuranceCoverageID:
		return LifeInsuranceCoverage(), nil
	case NetWorthID:
		return NetWorth(), nil
	case OtherID:
		return Other(), nil
	default:
		return Tag{}, errors.Wrapf(ErrInvalidTagID, "%v", tagID)
	}
}

func Income() Tag {
	return Tag{
		ID:          IncomeID,
		Name:        "Income",
		LongName:    "Household Income",
		Description: "Total pre-tax household income",
		SortOrder:   10,
	}
}

func EmergencySavings() Tag {
	return Tag{
		ID:          EmergencySavingsID,
		Name:        "Savings",
		LongName:    "Emergency Savings",
		Description: "Total household emergency savings",
		SortOrder:   20,
	}
}

func EducationSavings() Tag {
	return Tag{
		ID:          EducationSavingsID,
		Name:        "Education",
		LongName:    "Education Savings",
		Description: "Total household education savings",
		SortOrder:   30,
	}
}

func RetirementSavings() Tag {
	return Tag{
		ID:          RetirementSavingsID,
		Name:        "Retirement",
		LongName:    "Retirement Savings",
		Description: "Total household Retirement savings",
		SortOrder:   40,
	}
}

func LifeInsuranceCoverage() Tag {
	return Tag{
		ID:          LifeInsuranceCoverageID,
		Name:        "Insurance",
		LongName:    "Life Insurance",
		Description: "Total Life Insurance Coverage",
		SortOrder:   50,
	}
}

func NetWorth() Tag {
	return Tag{
		ID:          NetWorthID,
		Name:        "Net Worth",
		LongName:    "Net Worth",
		Description: "Total Household Net Worth",
		SortOrder:   60,
	}
}

func Other() Tag {
	return Tag{
		ID:          OtherID,
		Name:        "Other",
		LongName:    "Other",
		Description: "Other",
		SortOrder:   100,
	}
}
