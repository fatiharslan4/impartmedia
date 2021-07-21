package models

import (
	"time"
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
	HouseHold      string    `json:"household" `
	Dependents     string    `json:"dependents" `
	Generation     string    `json:"generation" `
	Gender         string    `json:"gender" `
	Race           string    `json:"race" `
	FinancialGoals string    `json:"financialGoals" `
}

type PageduserResponse struct {
	UserDetails UserDetails `json:"users"`
	NextPage    *NextPage   `json:"nextPage"`
}
