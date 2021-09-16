package plaid

import (
	"context"
	"fmt"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	plaid "github.com/plaid/plaid-go/plaid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (ser *plaidHandler) SavePlaidInstitutions(ctx context.Context) error {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", PLAID_CLIENT_ID)
	configuration.AddDefaultHeader("PLAID-SECRET", PLAID_SECRET)
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)
	var countrCode = []plaid.CountryCode{plaid.COUNTRYCODE_US}
	fmt.Println(client)
	request := plaid.NewInstitutionsGetRequest(Count, OffSet, countrCode)
	fmt.Println(request)
	resp, _, err := client.PlaidApi.InstitutionsGet(ctx).InstitutionsGetRequest(*request).Execute()
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, inst := range resp.Institutions {
		dbInstitution := ToDBModel(inst)
		err := dbInstitution.Insert(ctx, ser.db, boil.Infer())
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (ser *plaidHandler) SavePlaidInstitutionToken(ctx context.Context, userInstitution UserInstitution) error {

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", PLAID_CLIENT_ID)
	configuration.AddDefaultHeader("PLAID-SECRET", PLAID_SECRET)
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)
	var countrCode = []plaid.CountryCode{plaid.COUNTRYCODE_US}
	request := plaid.NewInstitutionsGetByIdRequest(userInstitution.PlaidInstitutionId, countrCode)
	response, _, err := client.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(*request).Execute()
	if err != nil {
		return err
	}
	out := &dbmodels.Institution{
		PlaidInstitutionID: response.Institution.InstitutionId,
		InstitutionName:    response.Institution.Name,
	}
	err = out.Insert(ctx, ser.db, boil.Infer())
	inst, err := dbmodels.Institutions(dbmodels.InstitutionWhere.PlaidInstitutionID.EQ(userInstitution.PlaidInstitutionId)).One(ctx, ser.db)
	if err != nil {
		fmt.Println(err)
	}
	userInstitution.CreatedAt = impart.CurrentUTC()
	userInstitution.Id = inst.ID
	dbUserInstitution := userInstitution.ToDBModel()
	err = dbUserInstitution.Insert(ctx, ser.db, boil.Infer())
	if err != nil {
		return err
	}
	return nil

}

func (ser *plaidHandler) GetPlaidInstitutions(ctx context.Context) (Institutions, error) {
	institutions, err := dbmodels.Institutions().All(ctx, ser.db)
	if err != nil {
		return nil, err
	}
	output := DBmodelsToResult(institutions)
	return output, nil

}

func (ser *plaidHandler) SavePlaidInstitutionAccounts(ctx context.Context, userInstitution UserInstitution) error {
	userInstitution.CreatedAt = impart.CurrentUTC()
	dbUserInstitution := userInstitution.ToDBModel()
	err := dbUserInstitution.Insert(ctx, ser.db, boil.Infer())
	if err != nil {
		return err
	}
	return nil

}

func (ser *plaidHandler) GetPlaidUserInstitutions(ctx context.Context, impartWealthId string) (UserInstitutions, error) {
	userInstitutions, err := dbmodels.UserInstitutions(dbmodels.UserInstitutionWhere.ImpartWealthID.EQ(impartWealthId),
		qm.Offset(0),
		qm.Limit(100),
		qm.Load(dbmodels.UserInstitutionRels.ImpartWealth),
		qm.Load(dbmodels.UserInstitutionRels.Institution),
	).All(ctx, ser.db)

	if err != nil {
		return nil, err
	}

	output := DBmodelsToUserInstitutionResult(userInstitutions)
	return output, nil
}

func (ser *plaidHandler) SavePlaidUserInstitutionAccpunts(ctx context.Context) error {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", PLAID_CLIENT_ID)
	configuration.AddDefaultHeader("PLAID-SECRET", PLAID_SECRET)
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)

	userInstitutions, err := dbmodels.UserInstitutions().All(ctx, ser.db)
	if err != nil {
		fmt.Println(err)
	}
	for _, user := range userInstitutions {
		accountsGetRequest := plaid.NewAccountsGetRequest(user.AccessToken)
		accountsGetRequest.SetOptions(plaid.AccountsGetRequestOptions{
			AccountIds: &[]string{},
		})
		accountsGetResp, _, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
			*accountsGetRequest,
		).Execute()
		if err != nil {
			fmt.Println(err)
		}
		accounts := accountsGetResp.GetAccounts()
		for _, act := range accounts {
			// dbAccnt := ToDBModel(act)
			fmt.Println(act)
		}
	}
	return nil
}
