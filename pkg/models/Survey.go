package models

import (
	"time"

	r "github.com/Pallinder/go-randomdata"
	"github.com/impartwealthapp/backend/pkg/impart"
)

type RawResponses struct {
	RespondentID   string `index:"0"`
	StartTimestamp string `index:"2"`
	EndTimestamp   string `index:"3"`

	BirthYear               string `index:"5"`
	Gender                  string `index:"6"`
	RelationshipStatus      string `index:"7"`
	EmploymentStatus        string `index:"8"`
	Location                string `index:"9"`
	HasChildrenUnder18      string `index:"11"`
	NumberOfChildrenUnder18 string `index:"16"`

	HouseholdIncomeAmount string `index:"31"`
	NetWorthAmount        string `index:"33"`

	TaxFilingStatus        string `index:"34"`
	MarginalTaxBracket     string `index:"35"`
	EmergencySavingsAmount string `index:"36"`

	SavingForEducation     string `index:"37"`
	EducationSavingsGoal   string `index:"39"`
	EducationSavingsAmount string `index:"40"`

	RetirementSavingsGoal   string `index:"49"`
	RetirementSavingsAmount string `index:"50"`

	HaveLifeInsurance   string `index:"57"`
	LifeInsuranceAmount string `index:"60"`
	HaveWill            string `index:"61"`
	Email               string `index:"66" conform:"email,lowercase"`
}

type SurveyResponses struct {
	Email                   string    `json:"email"`
	RespondentID            int64     `json:"respondentID,omitempty"`
	StartTimestamp          time.Time `json:"startTimestamp,omitempty"`
	EndTimestamp            time.Time `json:"endTimestamp,omitempty"`
	ImportTimestamp         time.Time `json:"importTimestamp,omitempty"`
	BirthYear               int       `json:"birthYear,omitempty"`
	Gender                  string    `json:"gender,omitempty"`
	RelationshipStatus      string    `json:"relationshipStatus,omitempty"`
	EmploymentStatus        string    `json:"employmentStatus,omitempty"`
	Location                string    `json:"location,omitempty"`
	NumberOfChildrenUnder18 int       `json:"numberOfChildrenUnder18,omitempty"`
	HouseholdIncomeAmount   int       `json:"householdIncomeAmount,omitempty"`
	NetWorthAmount          int       `json:"netWorthAmount,omitempty,omitempty"`
	TaxFilingStatus         string    `json:"taxFilingStatus,omitempty"`
	MarginalTaxBracket      int       `json:"marginalTaxBracket,omitempty"`
	EmergencySavingsAmount  int       `json:"emergencySavingsAmount,omitempty"`
	SavingForEducation      bool      `json:"savingForEducation,omitempty"`
	EducationSavingsGoal    int       `json:"educationSavingsGoal,omitempty"`
	EducationSavingsAmount  int       `json:"educationSavingsAmount,omitempty"`
	RetirementSavingsGoal   int       `json:"retirementSavingsGoal,omitempty"`
	RetirementSavingsAmount int       `json:"retirementSavingsAmount,omitempty"`
	LifeInsuranceAmount     int       `json:"lifeInsuranceAmount,omitempty"`
	HaveWill                bool      `json:"haveWill,omitempty"`
}

func RandomSurveyResponses() SurveyResponses {
	return SurveyResponses{
		ImportTimestamp:         impart.CurrentUTC(),
		BirthYear:               r.Number(1900, 2005),
		Gender:                  r.Noun(),
		RelationshipStatus:      r.Adjective(),
		EmploymentStatus:        r.Adjective(),
		Location:                r.Country(r.FullCountry),
		NumberOfChildrenUnder18: r.Number(0, 5),
		HouseholdIncomeAmount:   r.Number(35000, 200000),
		NetWorthAmount:          r.Number(0, 10000000),
		TaxFilingStatus:         r.Adjective(),
		MarginalTaxBracket:      r.Number(0, 35),
		EmergencySavingsAmount:  r.Number(0, 10000000),
		SavingForEducation:      r.Boolean(),
		EducationSavingsGoal:    r.Number(0, 10000000),
		EducationSavingsAmount:  r.Number(0, 10000000),
		RetirementSavingsAmount: r.Number(0, 10000000),
		RetirementSavingsGoal:   r.Number(0, 10000000),
		LifeInsuranceAmount:     r.Number(0, 10000000),
		HaveWill:                r.Boolean(),
	}
}
