package plaid

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	plaid "github.com/plaid/plaid-go/plaid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

func (ser *service) SavePlaidInstitutions(ctx context.Context) error {
	configuration := plaid.NewConfiguration()
	cfg, _ := config.GetImpart()
	if cfg != nil {
		configuration.AddDefaultHeader("PLAID-CLIENT-ID", cfg.PlaidClientId)
		configuration.AddDefaultHeader("PLAID-SECRET", cfg.PlaidSecret)
	}

	if cfg.Env == config.Production {
		configuration.UseEnvironment(plaid.Production)
	} else if cfg.Env == config.Preproduction {
		configuration.UseEnvironment(plaid.Development)
	} else {
		configuration.UseEnvironment(plaid.Sandbox)
	}

	client := plaid.NewAPIClient(configuration)
	var countrCode = []plaid.CountryCode{plaid.COUNTRYCODE_US}
	request := plaid.NewInstitutionsGetRequest(Count, OffSet, countrCode)
	resp, _, err := client.PlaidApi.InstitutionsGet(ctx).InstitutionsGetRequest(*request).Execute()
	if err != nil {
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
	cfg, _ := config.GetImpart()
	if cfg != nil {
		configuration.AddDefaultHeader("PLAID-CLIENT-ID", cfg.PlaidClientId)
		configuration.AddDefaultHeader("PLAID-SECRET", cfg.PlaidSecret)
	}
	if cfg.Env == config.Production {
		configuration.UseEnvironment(plaid.Production)
	} else if cfg.Env == config.Preproduction {
		configuration.UseEnvironment(plaid.Development)
	} else {
		configuration.UseEnvironment(plaid.Sandbox)
	}

	client := plaid.NewAPIClient(configuration)
	var countrCode = []plaid.CountryCode{plaid.COUNTRYCODE_US}
	var includeOptionalMetadata bool = true
	request := plaid.NewInstitutionsGetByIdRequest(userInstitution.PlaidInstitutionId, countrCode)
	data := plaid.NewInstitutionsGetByIdRequestOptions()
	request.Options = data
	request.Options.IncludeOptionalMetadata = &includeOptionalMetadata
	response, _, err := client.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(*request).Execute()
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find Plaid institution for the user.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", userInstitution.ImpartWealthID),
			zap.String("PlaidInstitutionId", userInstitution.PlaidInstitutionId))

		return impartErr
	}
	noInstitution := false
	inst, err := dbmodels.Institutions(dbmodels.InstitutionWhere.PlaidInstitutionID.EQ(userInstitution.PlaidInstitutionId)).One(ctx, ser.db)
	if err != nil {

		if err == sql.ErrNoRows {
			url := ""
			noInstitution = true
			if response.Institution.GetLogo() != "" {
				files := make([]models.File, 1)
				files[0].Content = response.Institution.GetLogo()
				files[0].FileName = "filename.png"
				files[0].FileType = "image/png"
				files = ValidatePostFilesName(ctx, files, response.Institution.InstitutionId, userInstitution.ImpartWealthID)
				postFiles, _ := ser.Hive.AddPostFiles(ctx, files)
				if len(postFiles) > 0 {
					url = postFiles[0].URL
				}
				// } else {
				// 	url = "https://impart-wealth-data-source-dev.s3.us-east-2.amazonaws.com/post/Adminone1/1631862956_Adminone1_filename.png"
				// }
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
	if !noInstitution {
		userInst, err := dbmodels.UserInstitutions(dbmodels.UserInstitutionWhere.ImpartWealthID.EQ(userInstitution.ImpartWealthID),
			dbmodels.UserInstitutionWhere.InstitutionID.EQ(inst.ID)).One(ctx, ser.db)
		if err != nil {
			ser.logger.Error("UserInstitutions fetching failed", zap.String("User", userInstitution.ImpartWealthID),
				zap.String("PlaidInstitutionId", userInstitution.PlaidInstitutionId),
				zap.Any("InstitutionID", inst.ID))
		}
		if userInst != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, "You already have an account with this bank, please register with some other bank account, thanks.")
			ser.logger.Error("Acces token saving failed", zap.String("User", userInstitution.ImpartWealthID),
				zap.Any("impartErr", impartErr))

			return impartErr
		}
	}
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

	output := DBmodelsToUserInstitutionResult(userInstitutions, ctx)
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

	if len(userInstitutions) == 0 {
		return UserAccount{}, impart.NewError(impart.ErrBadRequest, "No records found.")
	}
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user institution details.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserAccount{}, impartErr
	}

	configuration := plaid.NewConfiguration()
	cfg, _ := config.GetImpart()
	if cfg != nil {
		configuration.AddDefaultHeader("PLAID-CLIENT-ID", cfg.PlaidClientId)
		configuration.AddDefaultHeader("PLAID-SECRET", cfg.PlaidSecret)

		if cfg.Env == config.Production {
			configuration.UseEnvironment(plaid.Production)
		} else if cfg.Env == config.Preproduction {
			configuration.UseEnvironment(plaid.Development)
		} else {
			configuration.UseEnvironment(plaid.Sandbox)
		}

	}
	client := plaid.NewAPIClient(configuration)

	userData := UserAccount{}
	userData.ImpartWealthID = impartWealthId
	userData.UpdatedAt = time.Now().UTC().Unix()
	userinstitution := make(UserInstitutions, len(userInstitutions))
	finalQuery := ""
	for i, user := range userInstitutions {
		institution := InstitutionToModel(user)
		accountsGetRequest := plaid.NewAccountsGetRequest(user.AccessToken)
		accountsGetResp, response, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
			*accountsGetRequest,
		).Execute()

		if response.StatusCode == 400 {
			defer response.Body.Close()
			bodyBytes, _ := ioutil.ReadAll(response.Body)
			type errorResponse struct {
				ErrorCode string `json:"error_code" `
			}
			newRes := errorResponse{}
			err = json.Unmarshal(bodyBytes, &newRes)
			if err != nil {
				fmt.Println(err)
			}
			if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
				institution.IsAuthenticationError = true
			}
		}
		if err != nil {
			ser.logger.Error("Could not find the user plaid account details.", zap.String("User", impartWealthId),
				zap.String("token", user.AccessToken))
			continue
		}
		accounts := accountsGetResp.GetAccounts()
		userAccounts := make(Accounts, len(accounts))
		qury := ""
		query := ""
		logwrite := false
		for i, act := range accounts {
			userAccounts[i], qury = AccountToModel(act, user.UserInstitutionID)
			query = fmt.Sprintf("%s %s", query, qury)
			logwrite = true
		}
		institution.Accounts = userAccounts
		userinstitution[i] = institution
		userData.Institutions = userinstitution

		if logwrite {
			finalQuery = fmt.Sprintf("%s %s", finalQuery, query)
		}
	}
	if finalQuery != "" {
		go func() {
			lastQury := "INSERT INTO `user_plaid_accounts_log` (`user_institution_id`,`account_id`,`mask`,`name`,`official_name`,`subtype`,`type`,`iso_currency_code`,`unofficial_currency_code`,`available`,`current`,`credit_limit`,`created_at`) VALUES "
			lastQury = fmt.Sprintf("%s %s", lastQury, finalQuery)
			lastQury = strings.Trim(lastQury, ",")
			lastQury = fmt.Sprintf("%s ;", lastQury)
			ser.logger.Info("Query", zap.Any("Query", lastQury))
			_, err = queries.Raw(lastQury).QueryContext(ctx, ser.db)
			if err != nil {
				ser.logger.Error("error attempting to  log in user_plaid_accounts_log ", zap.Any("user_plaid_accounts_log", lastQury), zap.Error(err))
			}
		}()
	}
	return userData, nil
}

func AccountToModel(act plaid.AccountBase, userInstId uint64) (Account, string) {
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

	query := fmt.Sprintf("(%d,'%s','%s','%s','%s','%s','%s','%s','%s',%f,%f,%f,UTC_TIMESTAMP(3)),",
		userInstId, accounts.AccountID, accounts.Mask, accounts.Name, accounts.OfficialName, accounts.Subtype, accounts.Type, accounts.IsoCurrencyCode, accounts.UnofficialCurrencyCode, accounts.Available, accounts.Current, accounts.CreditLimit)

	return accounts, query
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
	if institution.Weburl != "" {
		weburl, err := url.Parse(institution.Weburl)
		if err == nil {
			institution.Weburl = weburl.Host
			// host, _, _ := net.SplitHostPort(weburl.Host)
		}
	}
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
	}
	return postFiles
}

func (ser *service) GetPlaidUserInstitutionTransactions(ctx context.Context, impartWealthId string, gpi models.GetPlaidInput) (UserTransaction, []PlaidError) {

	var newPlaidErr []PlaidError
	plaidErr := PlaidError{Error: "unable to complete the request",
		Msg:                 "",
		AuthenticationError: false}

	_, err := dbmodels.Users(dbmodels.UserWhere.ImpartWealthID.EQ(impartWealthId)).One(ctx, ser.db)
	if err != nil {
		plaidErr.Msg = "Could not find the user."
		// plaidErr.AccessToken = userInstitutions.AccessToken
		newPlaidErr = append(newPlaidErr, plaidErr)

		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserTransaction{}, newPlaidErr
	}

	userInstitutions, err := dbmodels.UserInstitutions(dbmodels.UserInstitutionWhere.ImpartWealthID.EQ(impartWealthId),
		qm.Load(dbmodels.UserInstitutionRels.ImpartWealth),
		qm.Load(dbmodels.UserInstitutionRels.Institution),
	).One(ctx, ser.db)

	if userInstitutions == nil {
		plaidErr.Msg = "No records found."
		plaidErr.AccessToken = userInstitutions.AccessToken
		newPlaidErr = append(newPlaidErr, plaidErr)
		return UserTransaction{}, newPlaidErr
	}
	if err != nil {
		plaidErr.Msg = "Could not find the user institution details."
		plaidErr.AccessToken = userInstitutions.AccessToken
		newPlaidErr = append(newPlaidErr, plaidErr)

		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserTransaction{}, newPlaidErr
	}

	configuration := plaid.NewConfiguration()
	cfg, _ := config.GetImpart()
	if cfg != nil {
		configuration.AddDefaultHeader("PLAID-CLIENT-ID", cfg.PlaidClientId)
		configuration.AddDefaultHeader("PLAID-SECRET", cfg.PlaidSecret)

		if cfg.Env == config.Production {
			configuration.UseEnvironment(plaid.Production)
		} else if cfg.Env == config.Preproduction {
			configuration.UseEnvironment(plaid.Development)
		} else {
			configuration.UseEnvironment(plaid.Sandbox)
		}

	}
	client := plaid.NewAPIClient(configuration)

	transGetRequest := plaid.NewTransactionsGetRequest(userInstitutions.AccessToken, impart.CurrentUTC().AddDate(0, 0, -30).Format("2006-01-02"), impart.CurrentUTC().Format("2006-01-02"))

	// var count int32 = 10
	// var offset int32 = 0
	data := plaid.NewTransactionsGetRequestOptions()
	transGetRequest.Options = data
	transGetRequest.Options.Count = &gpi.Limit
	transGetRequest.Options.Offset = &gpi.Offset

	transGetResp, resp, err := client.PlaidApi.TransactionsGet(ctx).TransactionsGetRequest(
		*transGetRequest,
	).Execute()

	if err != nil || resp.StatusCode == 400 {
		ser.logger.Error("Could not find the user plaid account details.", zap.String("User", impartWealthId),
			zap.String("token", userInstitutions.AccessToken))
		plaidErr.Msg = "Could not find the  transaction details."
		if resp.StatusCode == 400 {
			defer resp.Body.Close()
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			type errorResponse struct {
				ErrorCode string `json:"error_code" `
			}
			newRes := errorResponse{}
			err := json.Unmarshal(bodyBytes, &newRes)
			if err != nil {
				ser.logger.Error("Could not unmarshal bodyBytes.", zap.Any("bodyBytes", bodyBytes),
					zap.String("token", userInstitutions.AccessToken))
			}
			if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
				plaidErr.AuthenticationError = true
				plaidErr.Msg = "ITEM_LOGIN_REQUIRED"
			}
			plaidErr.AccessToken = userInstitutions.AccessToken
		}
		newPlaidErr = append(newPlaidErr, plaidErr)

		// impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the  transaction details.")
		return UserTransaction{}, newPlaidErr
	}
	transactions := transGetResp.GetTransactions()
	if len(transactions) == 0 {
		plaidErr.Msg = "Could not find the  transaction details."
		plaidErr.AccessToken = userInstitutions.AccessToken
		newPlaidErr = append(newPlaidErr, plaidErr)

		// impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user transaction details.")
		return UserTransaction{}, newPlaidErr
	}
	userData := UserTransaction{}
	userData.ImpartWealthID = impartWealthId
	userData.AccessToken = userInstitutions.AccessToken
	userinstitution := make(UserInstitutions, 1)
	var transDatawithdateFinalData []Transaction
	var transDataFinalData []TransactionWithDate
	institution := InstitutionToModel(userInstitutions)

	var allDates []string
	fmt.Println("len(transactions)")
	fmt.Println(len(transactions))
	for _, act := range transactions {
		currentDate := act.Date

		ser.logger.Info(currentDate)
		ser.logger.Info("allDates", zap.Any("allDates", allDates))
		if !checkDateExist(currentDate, allDates) {
			ser.logger.Info("alredy date added", zap.Any("allDates", allDates),
				zap.Any("currentDate", currentDate))

			for _, acnts := range transactions {
				if currentDate == acnts.Date {
					ser.logger.Info("acnts.Date", zap.Any("acnts.Date", acnts.Date))
					if !checkDateExist(currentDate, allDates) {
						allDates = append(allDates, currentDate)
					}
					ser.logger.Info("aallDates", zap.Any("allDates", allDates))
					transDatawithdate := TransactionToModel(acnts, userInstitutions.UserInstitutionID)
					transDatawithdateFinalData = append(transDatawithdateFinalData, transDatawithdate)
				}
			}
			transWIthdate := TransactionWithDate{}
			transWIthdate.Date = currentDate
			transWIthdate.Data = transDatawithdateFinalData
			transDataFinalData = append(transDataFinalData, transWIthdate)
			transDatawithdateFinalData = nil
		}
	}

	userinstitution[0] = institution
	userData.Transactions = transDataFinalData
	userData.TotalTransaction = transGetResp.GetTotalTransactions()
	return userData, nil
}

func TransactionToModel(act plaid.Transaction, userInstId uint64) Transaction {
	trans := Transaction{}
	trans.AccountID = act.AccountId
	trans.Amount = act.GetAmount()
	trans.Category = act.Category
	trans.Name = act.Name
	trans.Date = act.GetDate()
	return trans
}

func checkDateExist(datenew string, alldate []string) bool {
	for _, date := range alldate {
		if date == datenew {
			return true
		}
	}
	return false
}

func GetAccessTokenStatus(accessToken string, ctx context.Context) bool {
	configuration := plaid.NewConfiguration()
	cfg, _ := config.GetImpart()
	if cfg != nil {
		configuration.AddDefaultHeader("PLAID-CLIENT-ID", cfg.PlaidClientId)
		configuration.AddDefaultHeader("PLAID-SECRET", cfg.PlaidSecret)
		if cfg.Env == config.Production {
			configuration.UseEnvironment(plaid.Production)
		} else if cfg.Env == config.Preproduction {
			configuration.UseEnvironment(plaid.Development)
		} else {
			configuration.UseEnvironment(plaid.Sandbox)
		}
	}
	client := plaid.NewAPIClient(configuration)
	accountsGetRequest := plaid.NewAccountsGetRequest(accessToken)
	_, response, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*accountsGetRequest,
	).Execute()
	if response.StatusCode == 400 {
		defer response.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		type errorResponse struct {
			ErrorCode string `json:"error_code" `
		}
		newRes := errorResponse{}
		err = json.Unmarshal(bodyBytes, &newRes)
		if err != nil {
			fmt.Println(err)
		}
		if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
			return true
		}
	}
	return false
}
