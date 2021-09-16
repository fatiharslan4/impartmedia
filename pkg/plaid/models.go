package plaid

import (
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	plaid "github.com/plaid/plaid-go/plaid"
	"github.com/volatiletech/null/v8"
)

type Institutions []Institution
type Institution struct {
	ID                 uint64      `json:"Id" `
	PlaidInstitutionId string      `json:"plaid_institution_id"`
	InstitutionName    string      `json:"institution_name"`
	Logo               null.String `json:"logo"`
	Weburl             string      `json:"weburl"`
	RequestId          string      `json:"request_id"`
}

type UserInstitutions []UserInstitution
type UserInstitution struct {
	UserInstitutionsId uint64    `json:"user_institutions_id" `
	Id                 uint64    `json:"id"`
	ImpartWealthID     string    `json:"impartWealthId" `
	AccessToken        string    `json:"access_token"`
	CreatedAt          time.Time `json:"created_at"`
	PlaidInstitutionId string    `json:"plaid_institution_id"`
}

type Balance struct {
	Available              uint64      `json:"available" `
	Current                uint64      `json:"current"`
	IsoCurrencyCode        string      `json:"iso_currency_code" `
	CreditLimit            null.Uint64 `json:"credit_limit"`
	UnofficialCurrencyCode null.String `json:"unofficial_currency_code"`
}

type UserInstitutionAccounts []UserInstitutionAccount
type UserInstitutionAccount struct {
	AccountId          uint64  `json:"account_id" `
	UserInstitutionsId uint64  `json:"user_institutions_id" `
	Balance            Balance `json:"balance"`
	Mask               string  `json:"mask" `
	Name               string  `json:"name"`
	OfficialName       string  `json:"official_name"`
	Subtype            string  `json:"subtype"`
	Type               string  `json:"type"`
	PlaidAccountId     string  `json:"plaid_account_id"`
}

func ToDBModel(p plaid.Institution) *dbmodels.Institution {
	out := &dbmodels.Institution{
		PlaidInstitutionID: p.InstitutionId,
		InstitutionName:    p.Name,
	}

	return out
}

func DBmodelsToResult(dbInstitution dbmodels.InstitutionSlice) Institutions {
	out := make(Institutions, len(dbInstitution))
	for i, p := range dbInstitution {
		out[i] = InstitutionFromDB(p)
	}
	return out
}

func InstitutionFromDB(p *dbmodels.Institution) Institution {
	out := Institution{
		ID:                 p.ID,
		PlaidInstitutionId: p.PlaidInstitutionID,
		InstitutionName:    p.InstitutionName,
	}

	return out
}

func (p UserInstitution) ToDBModel() *dbmodels.UserInstitution {
	out := &dbmodels.UserInstitution{
		ImpartWealthID: p.ImpartWealthID,
		CreatedAt:      p.CreatedAt,
		AccessToken:    p.AccessToken,
		InstitutionID:  p.Id,
	}

	return out
}

func DBmodelsToUserInstitutionResult(dbInstitution dbmodels.UserInstitutionSlice) UserInstitutions {
	out := make(UserInstitutions, len(dbInstitution))
	for i, p := range dbInstitution {
		out[i] = UserInstitutionFromDB(p)
	}
	return out
}

func UserInstitutionFromDB(p *dbmodels.UserInstitution) UserInstitution {
	out := UserInstitution{
		AccessToken:        p.AccessToken,
		CreatedAt:          p.CreatedAt,
		Id:                 p.InstitutionID,
		PlaidInstitutionId: p.R.Institution.PlaidInstitutionID,
		ImpartWealthID:     p.ImpartWealthID,
		UserInstitutionsId: p.UserInstitutionsID,
	}

	return out
}

type NextPage struct {
	Offset int `json:"offset"`
}

type PagedUserInstitutionResponse struct {
	Userinstitution UserInstitutions `json:"usernstitution"`
	NextPage        *NextPage        `json:"nextPage"`
}

type PagedInstitutionResponse struct {
	Institution Institutions `json:"userinstitution"`
	NextPage    *NextPage    `json:"nextPage"`
}

// func AccountToDBModel(p plaid.AccountBase) *dbmodels.UserInstitutionAccount {
// 	out := &dbmodels.UserInstitutionAccount{
// 		PlaidAccountID: p.AccountId,
// 		// Mask:           p.Mask,
// 		Name: p.Name,
// 		// OfficialName: p.OfficialName,
// 		// Subtype: p.Subtype,
// 		// Subtype: p.Type,
// 	}
// 	val, _ := p.Balances.Limit.MarshalJSON()
// 	out.CreditLimit = val
// 	return out
// }
