package plaid

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	plaid "github.com/plaid/plaid-go/plaid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

func (ser *service) SavePlaidInstitutions(ctx context.Context) error {
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

func (ser *service) SavePlaidInstitutionToken(ctx context.Context, userInstitution UserInstitutionToken) impart.Error {

	_, err := dbmodels.Users(dbmodels.UserWhere.ImpartWealthID.EQ(userInstitution.ImpartWealthID)).One(ctx, ser.db)
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", userInstitution.ImpartWealthID),
			zap.String("user", userInstitution.ImpartWealthID))
		return impartErr
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", PLAID_CLIENT_ID)
	configuration.AddDefaultHeader("PLAID-SECRET", PLAID_SECRET)
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)
	var countrCode = []plaid.CountryCode{plaid.COUNTRYCODE_US}
	var includeOptionalMetadata bool = true
	request := plaid.NewInstitutionsGetByIdRequest(userInstitution.PlaidInstitutionId, countrCode)
	data := plaid.NewInstitutionsGetByIdRequestOptions()
	request.Options = data
	request.Options.IncludeOptionalMetadata = &includeOptionalMetadata
	response, _, err := client.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(*request).Execute()
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not Plaid institution the user.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", userInstitution.ImpartWealthID),
			zap.String("PlaidInstitutionId", userInstitution.PlaidInstitutionId))

		return impartErr
	}
	inst, err := dbmodels.Institutions(dbmodels.InstitutionWhere.PlaidInstitutionID.EQ(userInstitution.PlaidInstitutionId)).One(ctx, ser.db)
	if err != nil {

		if err == sql.ErrNoRows {
			url := ""
			if response.Institution.GetLogo() != "" {
				files := make([]models.File, 1)
				files[0].Content = response.Institution.GetLogo()
				files[0].FileName = "filename.png"
				files[0].FileType = "image/png"
				files = ValidatePostFilesName(ctx, files, response.Institution.InstitutionId, userInstitution.ImpartWealthID)
				postFiles, _ := ser.Hive.AddPostFiles(ctx, files)
				if len(postFiles) > 0 {
					url = postFiles[0].URL
				} else {
					url = "https://impart-wealth-data-source-dev.s3.us-east-2.amazonaws.com/post/Adminone1/1631862956_Adminone1_filename.png"
				}
			}
			out := &dbmodels.Institution{
				PlaidInstitutionID: response.Institution.InstitutionId,
				InstitutionName:    response.Institution.Name,
				Logo:               url,
				Weburl:             response.Institution.GetUrl(),
			}
			err = out.Insert(ctx, ser.db, boil.Infer())
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "Institution save failed.")
				ser.logger.Error("Institution save failed.", zap.String("User", userInstitution.ImpartWealthID),
					zap.String("PlaidInstitutionId", userInstitution.PlaidInstitutionId))

				return impartErr
			}
			inst, err = dbmodels.Institutions(dbmodels.InstitutionWhere.PlaidInstitutionID.EQ(userInstitution.PlaidInstitutionId)).One(ctx, ser.db)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, "Institution fetching failed.")
				ser.logger.Error("IInstitution fetching failed in db.", zap.String("User", userInstitution.ImpartWealthID),
					zap.String("PlaidInstitutionId", userInstitution.PlaidInstitutionId))

				return impartErr
			}
		}
	}
	userInstitution.CreatedAt = impart.CurrentUTC()
	userInstitution.Id = inst.ID
	dbUserInstitution := userInstitution.ToDBModel()
	err = dbUserInstitution.Insert(ctx, ser.db, boil.Infer())
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Acces token saving failed.")
		ser.logger.Error("Acces token saving failed", zap.String("User", userInstitution.ImpartWealthID),
			zap.String("PlaidInstitutionId", userInstitution.PlaidInstitutionId))

		return impartErr
	}
	return nil

}

func (ser *service) GetPlaidInstitutions(ctx context.Context) (Institutions, error) {
	institutions, err := dbmodels.Institutions().All(ctx, ser.db)
	if err != nil {
		return nil, err
	}
	output := DBmodelsToResult(institutions)
	return output, nil

}

func (ser *service) GetPlaidUserInstitutions(ctx context.Context, impartWealthId string) (UserInstitutionTokens, error) {
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

func (ser *service) GetPlaidUserInstitutionAccounts(ctx context.Context, impartWealthId string) (UserAccount, impart.Error) {

	_, err := dbmodels.Users(dbmodels.UserWhere.ImpartWealthID.EQ(impartWealthId)).One(ctx, ser.db)
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserAccount{}, impartErr
	}

	userInstitutions, err := dbmodels.UserInstitutions(dbmodels.UserInstitutionWhere.ImpartWealthID.EQ(impartWealthId),
		qm.Load(dbmodels.UserInstitutionRels.ImpartWealth),
		qm.Load(dbmodels.UserInstitutionRels.Institution),
	).All(ctx, ser.db)

	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user institution details.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserAccount{}, impartErr
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", PLAID_CLIENT_ID)
	configuration.AddDefaultHeader("PLAID-SECRET", PLAID_SECRET)
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)

	userData := UserAccount{}
	userData.ImpartWealthID = impartWealthId
	userData.UpdatedAt = impart.CurrentUTC()
	userinstitution := make(UserInstitutions, len(userInstitutions))
	for i, user := range userInstitutions {
		institution := InstitutionToModel(user)
		accountsGetRequest := plaid.NewAccountsGetRequest(user.AccessToken)
		accountsGetResp, _, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
			*accountsGetRequest,
		).Execute()
		if err != nil {
			ser.logger.Error("Could not find the user plaid account details.", zap.String("User", impartWealthId),
				zap.String("token", user.AccessToken))
			continue
		}
		accounts := accountsGetResp.GetAccounts()
		userAccounts := make(Accounts, len(accounts))
		for i, act := range accounts {
			userAccounts[i] = AccountToModel(act)
		}
		institution.Accounts = userAccounts
		userinstitution[i] = institution
		userData.Institutions = userinstitution
	}
	return userData, nil
}

func AccountToModel(act plaid.AccountBase) Account {
	accounts := Account{}
	accounts.AccountID = act.AccountId
	accounts.Available = act.Balances.GetAvailable()
	accounts.Current = act.Balances.GetCurrent()
	accounts.CreditLimit = act.Balances.GetLimit()
	accounts.IsoCurrencyCode = act.Balances.GetIsoCurrencyCode()
	accounts.Mask = act.GetMask()
	accounts.Type = string(act.Type)
	accounts.Subtype = string(act.GetSubtype())
	accounts.Name = act.GetName()
	accounts.OfficialName = act.GetOfficialName()
	accounts.UnofficialCurrencyCode = act.Balances.GetUnofficialCurrencyCode()
	return accounts
}

func InstitutionToModel(user *dbmodels.UserInstitution) UserInstitution {
	institution := UserInstitution{}
	institution.AccessToken = user.AccessToken
	institution.PlaidInstitutionId = user.R.Institution.PlaidInstitutionID
	institution.CreatedAt = user.CreatedAt
	institution.Logo = user.R.Institution.Logo
	institution.Weburl = user.R.Institution.Weburl
	institution.ImpartWealthID = user.ImpartWealthID
	institution.InstitutionName = user.R.Institution.InstitutionName
	return institution
}

func ValidatePostFilesName(ctx context.Context, postFiles []models.File, institution_id string, impartWealthID string) []models.File {
	basePath := fmt.Sprintf("%s/", "plaid")
	pattern := `[^\[0-9A-Za-z_.-]`
	for index := range postFiles {
		filename := fmt.Sprintf("%d_%s_%s",
			time.Now().Unix(),
			institution_id,
			postFiles[index].FileName,
		)

		re, _ := regexp.Compile(pattern)
		filename = re.ReplaceAllString(filename, "")

		postFiles[index].FilePath = basePath
		postFiles[index].FileName = filename
		fmt.Println(basePath)
		fmt.Println(filename)
	}
	return postFiles
}
