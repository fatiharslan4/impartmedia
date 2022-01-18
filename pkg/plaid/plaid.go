package plaid

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	} else if cfg.Env == config.Development {
		configuration.UseEnvironment(plaid.Sandbox)
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
	} else if cfg.Env == config.Development { //test
		configuration.UseEnvironment(plaid.Sandbox)
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

func (ser *service) GetPlaidUserInstitutionAccounts(ctx context.Context, impartWealthId string, gpi models.GetPlaidInput) (UserAccount, *NextPage, impart.Error) {

	_, err := dbmodels.Users(dbmodels.UserWhere.ImpartWealthID.EQ(impartWealthId)).One(ctx, ser.db)
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserAccount{}, nil, impartErr
	}
	if gpi.Limit <= 0 {
		gpi.Limit = 100
	}

	userInstitutions, err := dbmodels.UserInstitutions(dbmodels.UserInstitutionWhere.ImpartWealthID.EQ(impartWealthId),
		qm.Load(dbmodels.UserInstitutionRels.ImpartWealth),
		qm.Load(dbmodels.UserInstitutionRels.Institution),
		qm.Limit(int(gpi.Limit)),
		qm.Offset(int(gpi.Offset)),
	).All(ctx, ser.db)

	if len(userInstitutions) == 0 {
		return UserAccount{}, nil, impart.NewError(impart.ErrBadRequest, "No records found.")
	}
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user institution details.")
		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserAccount{}, nil, impartErr
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
		} else if cfg.Env == config.Development {
			configuration.UseEnvironment(plaid.Sandbox)
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
	var totalAsset float32
	var acctCount int32
	for i, user := range userInstitutions {
		institution := InstitutionToModel(user)
		accountsGetRequest := plaid.NewAccountsGetRequest(user.AccessToken)
		accountsGetResp, response, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
			*accountsGetRequest,
		).Execute()

		if response.StatusCode == 400 {
			defer response.Body.Close()
			type errorResponse struct {
				ErrorCode string `json:"error_code" `
			}
			newRes := errorResponse{}
			json.NewDecoder(response.Body).Decode(&newRes)
			if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
				institution.IsAuthenticationError = true
			}
		}
		if err != nil {
			ser.logger.Error("Could not find the user plaid account details.", zap.String("User", impartWealthId),
				zap.String("token", user.AccessToken))
			// continue
		}
		accounts := accountsGetResp.GetAccounts()
		userAccounts := make(Accounts, len(accounts))
		qury := ""
		query := ""
		logwrite := false
		logexist := false
		userlogs, err := dbmodels.UserPlaidAccountsLogs(dbmodels.UserPlaidAccountsLogWhere.UserInstitutionID.EQ(user.UserInstitutionID)).One(ctx, ser.db)
		if userlogs != nil {
			logexist = true
		}
		if err != nil {
			fmt.Println(err)
			ser.logger.Error("error checking UserPlaidAccountsLogExists ", zap.Error(err))
		}

		for i, act := range accounts {
			if act.Type == "depository" || act.Type == "investment" || act.Type == "brokerage" {
				totalAsset += float32(act.Balances.GetCurrent())
				acctCount += 1
			}
			userAccounts[i], qury = AccountToModel(act, user.UserInstitutionID, logexist)
			if !logexist {
				query = fmt.Sprintf("%s %s", query, qury)
				logwrite = true
			}
		}
		institution.Accounts = userAccounts
		// institution.TotalAsset = totalAsset
		// institution.AccountCount = acctCount
		userinstitution[i] = institution
		userData.Institutions = userinstitution
		userData.TotalAsset = totalAsset
		userData.AccountCount = acctCount

		if logwrite {
			finalQuery = fmt.Sprintf("%s %s", finalQuery, query)
		}
	}
	if strings.Trim(finalQuery, "") != "" {
		go func() {
			tx, err := ser.db.BeginTx(ctx, nil)
			if err != nil {
				ser.logger.Error("Query", zap.Any("Query", err))
			} else {
				defer impart.CommitRollbackLogger(tx, err, ser.logger)
				lastQury := "LOCK TABLE user_plaid_accounts_log WRITE ;INSERT INTO `user_plaid_accounts_log` (`user_institution_id`,`account_id`,`mask`,`name`,`official_name`,`subtype`,`type`,`iso_currency_code`,`unofficial_currency_code`,`available`,`current`,`credit_limit`,`created_at`) VALUES "
				lastQury = fmt.Sprintf("%s %s", lastQury, finalQuery)
				lastQury = strings.Trim(lastQury, ",")
				lastQury = fmt.Sprintf("%s ; UNLOCK TABLES;", lastQury)
				_, err = queries.Raw(lastQury).ExecContext(ctx, ser.db)
				if err != nil {
					ser.logger.Error("error attempting to  log in user_plaid_accounts_log ", zap.Error(err))
				}
				tx.Commit()
			}
		}()
	}
	outOffset := &NextPage{
		Offset: int(gpi.Offset),
	}
	if len(userInstitutions) < int(gpi.Limit) {
		outOffset = nil
	} else {
		outOffset.Offset += len(userInstitutions)
	}
	return userData, outOffset, nil
}

func AccountToModel(act plaid.AccountBase, userInstId uint64, logexist bool) (Account, string) {
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
	accounts.DisplayValue = act.Balances.GetCurrent()
	accounts.DisplayName = act.GetName()
	query := ""
	if !logexist {
		query = fmt.Sprintf("(%d,'%s','%s','%s','%s','%s','%s','%s','%s',%f,%f,%f,UTC_TIMESTAMP(3)),",
			userInstId, accounts.AccountID, accounts.Mask, accounts.Name, accounts.OfficialName, accounts.Subtype, accounts.Type, accounts.IsoCurrencyCode, accounts.UnofficialCurrencyCode, accounts.Available, accounts.Current, accounts.CreditLimit)
	}

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
	institution.UserInstitutionsId = user.UserInstitutionID
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

func (ser *service) GetPlaidUserInstitutionTransactions(ctx context.Context, impartWealthId string, gpi models.GetPlaidInput) (UserTransaction, *NextPage, []PlaidError) {

	var totalTransaction int32

	var newPlaidErr []PlaidError
	plaidErr := PlaidError{Error: "unable to complete the request",
		Msg:                 "No transaction records found.",
		AuthenticationError: false}

	_, err := dbmodels.Users(dbmodels.UserWhere.ImpartWealthID.EQ(impartWealthId)).One(ctx, ser.db)
	if err != nil {
		plaidErr.Msg = "Could not find the user."
		// plaidErr.AccessToken = userInstitutions.AccessToken
		newPlaidErr = append(newPlaidErr, plaidErr)

		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserTransaction{}, nil, newPlaidErr
	}

	userInstitutionList, err := dbmodels.UserInstitutions(dbmodels.UserInstitutionWhere.ImpartWealthID.EQ(impartWealthId),
		qm.Load(dbmodels.UserInstitutionRels.ImpartWealth),
		qm.Load(dbmodels.UserInstitutionRels.Institution),
	).All(ctx, ser.db)
	if userInstitutionList == nil {
		plaidErr.Msg = "No records found."
		newPlaidErr = append(newPlaidErr, plaidErr)
		return UserTransaction{}, nil, newPlaidErr
	}
	if err != nil {
		plaidErr.Msg = "Could not find the user institution details."
		// plaidErr.AccessToken = userInstitutions.AccessToken
		newPlaidErr = append(newPlaidErr, plaidErr)

		ser.logger.Error("Could not find the user institution details.", zap.String("User", impartWealthId),
			zap.String("user", impartWealthId))
		return UserTransaction{}, nil, newPlaidErr
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
		} else if cfg.Env == config.Development {
			configuration.UseEnvironment(plaid.Sandbox)
		} else {
			configuration.UseEnvironment(plaid.Sandbox)
		}

	}
	client := plaid.NewAPIClient(configuration)
	userData := UserTransaction{}
	var investTransactions []plaid.InvestmentTransaction
	for _, userInstitutions := range userInstitutionList {
		if userInstitutions.BankType == 1 {
			transGetRequest := plaid.NewTransactionsGetRequest(userInstitutions.AccessToken, impart.CurrentUTC().AddDate(0, 0, -30).Format("2006-01-02"), impart.CurrentUTC().Format("2006-01-02"))
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
					type errorResponse struct {
						ErrorCode string `json:"error_code" `
					}
					newRes := errorResponse{}
					json.NewDecoder(resp.Body).Decode(&newRes)
					if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
						plaidErr.AuthenticationError = true
						plaidErr.Msg = "ITEM_LOGIN_REQUIRED"
					}
					plaidErr.AccessToken = userInstitutions.AccessToken
				}
				newPlaidErr = append(newPlaidErr, plaidErr)

				// impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the  transaction details.")
				return UserTransaction{}, nil, newPlaidErr
			}
			transactions := transGetResp.GetTransactions()
			if len(transactions) > 0 {
				totalTransaction = int32(len(transactions))
				userData.ImpartWealthID = impartWealthId
				userData.AccessToken = userInstitutions.AccessToken
				userinstitution := make(UserInstitutions, 1)
				var transDatawithdateFinalData []Transaction
				var transDataFinalData []TransactionWithDate
				institution := InstitutionToModel(userInstitutions)
				var allDates []string
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

				outOffset := &NextPage{
					Offset: int(gpi.Offset),
				}
				if totalTransaction < gpi.Limit {
					outOffset = nil
				} else {
					outOffset.Offset += int(totalTransaction)
				}
				return userData, outOffset, nil
			}

		} else if userInstitutions.BankType == 2 {
			transInvestGetRequest := plaid.NewInvestmentsTransactionsGetRequest(userInstitutions.AccessToken, impart.CurrentUTC().AddDate(0, 0, -30).Format("2006-01-02"), impart.CurrentUTC().Format("2006-01-02"))
			data := plaid.NewInvestmentsTransactionsGetRequestOptions()
			transInvestGetRequest.Options = data
			transInvestGetRequest.Options.Count = &gpi.Limit
			transInvestGetRequest.Options.Offset = &gpi.Offset
			transGetResp, resp, err := client.PlaidApi.InvestmentsTransactionsGet(ctx).InvestmentsTransactionsGetRequest(
				*transInvestGetRequest,
			).Execute()
			if err != nil || resp.StatusCode == 400 {
				ser.logger.Error("Could not find the user plaid account details.", zap.String("User", impartWealthId),
					zap.String("token", userInstitutions.AccessToken))
				plaidErr.Msg = "Could not find the  transaction details."
				if resp.StatusCode == 400 {
					defer resp.Body.Close()
					type errorResponse struct {
						ErrorCode string `json:"error_code" `
					}
					newRes := errorResponse{}
					json.NewDecoder(resp.Body).Decode(&newRes)
					if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
						plaidErr.AuthenticationError = true
						plaidErr.Msg = "ITEM_LOGIN_REQUIRED"
					}
					plaidErr.AccessToken = userInstitutions.AccessToken
				}
				newPlaidErr = append(newPlaidErr, plaidErr)

				// impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the  transaction details.")
				return UserTransaction{}, nil, newPlaidErr
			}
			investTransactions = transGetResp.GetInvestmentTransactions()
			totalTransaction = int32(len(investTransactions))
			if totalTransaction > 0 {
				userData.ImpartWealthID = impartWealthId
				userData.AccessToken = userInstitutions.AccessToken
				userinstitution := make(UserInstitutions, 1)
				var transDatawithdateFinalData []Transaction
				var transDataFinalData []TransactionWithDate
				institution := InstitutionToModel(userInstitutions)
				var allDates []string
				for _, act := range investTransactions {
					currentDate := act.Date
					ser.logger.Info(currentDate)
					ser.logger.Info("allDates", zap.Any("allDates", allDates))
					if !checkDateExist(currentDate, allDates) {
						ser.logger.Info("alredy date added", zap.Any("allDates", allDates),
							zap.Any("currentDate", currentDate))

						for _, acnts := range investTransactions {
							if currentDate == acnts.Date {
								ser.logger.Info("acnts.Date", zap.Any("acnts.Date", acnts.Date))
								if !checkDateExist(currentDate, allDates) {
									allDates = append(allDates, currentDate)
								}
								ser.logger.Info("aallDates", zap.Any("allDates", allDates))
								transDatawithdate := InvestmentTransactionToModel(acnts, userInstitutions.UserInstitutionID)
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
				userData.TotalTransaction = transGetResp.GetTotalInvestmentTransactions()

				outOffset := &NextPage{
					Offset: int(gpi.Offset),
				}

				if totalTransaction < gpi.Limit {
					outOffset = nil
				} else {
					outOffset.Offset += int(totalTransaction)
				}
				return userData, outOffset, nil
			}
		}

		// transactions := transGetResp.GetTransactions()
		// if len(transactions) == 0 {
		// 	isInvestments := false
		// 	accounts := transGetResp.GetAccounts()
		// 	for _, accnt := range accounts {
		// 		if accnt.Type != "investment" {
		// 			return UserTransaction{}, nil, nil
		// 		} else {
		// 			isInvestments = true
		// 		}
		// 	}
		// 	if isInvestments {
		// 		plaidErr.Msg = "The transaction is empty since investment transactions are not supported."
		// 		plaidErr.AccessToken = userInstitutions.AccessToken
		// 		newPlaidErr = append(newPlaidErr, plaidErr)
		// 		return UserTransaction{}, nil, newPlaidErr
		// 	}
		// 	plaidErr.Msg = "Could not find the  transaction details."
		// 	plaidErr.AccessToken = userInstitutions.AccessToken
		// 	newPlaidErr = append(newPlaidErr, plaidErr)

		// 	// impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user transaction details.")
		// 	return UserTransaction{}, nil, nil
		// }

	}

	// outOffset := &NextPage{
	// 	Offset: int(gpi.Offset),
	// }

	// if totalTransaction < gpi.Limit {
	// 	outOffset = nil
	// } else {
	// 	outOffset.Offset += int(totalTransaction)
	// }
	return UserTransaction{}, nil, newPlaidErr
}

func (ser *service) GetPlaidUserAccountsTransactions(ctx context.Context, accountId string, userInstId uint64, impartWealthId string, gpi models.GetPlaidAccountTransactionInput) (UserTransaction, *NextPage, []PlaidError) {

	var totalTransaction int32

	var newPlaidErr []PlaidError
	plaidErr := PlaidError{Error: "unable to complete the request",
		Msg:                 "No transaction records found.",
		AuthenticationError: false}
	userInstitutionList, err := dbmodels.UserInstitutions(dbmodels.UserInstitutionWhere.UserInstitutionID.EQ(userInstId),
		qm.Load(dbmodels.UserInstitutionRels.ImpartWealth),
		qm.Load(dbmodels.UserInstitutionRels.Institution),
	).All(ctx, ser.db)
	if userInstitutionList == nil {
		plaidErr.Msg = "No records found."
		newPlaidErr = append(newPlaidErr, plaidErr)
		return UserTransaction{}, nil, newPlaidErr
	}
	accessToken := userInstitutionList[0].AccessToken
	if err != nil {
		plaidErr.Msg = "Could not find the user institution details."
		// plaidErr.AccessToken = userInstitutions.AccessToken
		newPlaidErr = append(newPlaidErr, plaidErr)

		ser.logger.Error("Could not find the user institution details.", zap.String("access_token", accessToken),
			zap.String("access_token", accessToken))
		return UserTransaction{}, nil, newPlaidErr
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
		} else if cfg.Env == config.Development {
			configuration.UseEnvironment(plaid.Sandbox)
		} else {
			configuration.UseEnvironment(plaid.Sandbox)
		}

	}
	client := plaid.NewAPIClient(configuration)
	userData := UserTransaction{}
	var investTransactions []plaid.InvestmentTransaction
	if userInstitutionList[0].BankType == 1 {
		transGetRequest := plaid.NewTransactionsGetRequest(accessToken, impart.CurrentUTC().AddDate(0, 0, -30).Format("2006-01-02"), impart.CurrentUTC().Format("2006-01-02"))
		data := plaid.NewTransactionsGetRequestOptions()
		var accountIds []string
		accountIds = append(accountIds, accountId)
		transGetRequest.Options = data
		transGetRequest.Options.Count = &gpi.Limit
		transGetRequest.Options.Offset = &gpi.Offset
		transGetRequest.Options.AccountIds = &accountIds

		transGetResp, resp, err := client.PlaidApi.TransactionsGet(ctx).TransactionsGetRequest(
			*transGetRequest,
		).Execute()
		if err != nil || resp.StatusCode == 400 {
			ser.logger.Error("Could not find the user plaid account details.",
				zap.String("token", accessToken))
			plaidErr.Msg = "Could not find the  transaction details."
			if resp.StatusCode == 400 {
				defer resp.Body.Close()
				type errorResponse struct {
					ErrorCode string `json:"error_code" `
				}
				newRes := errorResponse{}
				json.NewDecoder(resp.Body).Decode(&newRes)
				if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
					plaidErr.AuthenticationError = true
					plaidErr.Msg = "ITEM_LOGIN_REQUIRED"
				}
				plaidErr.AccessToken = accessToken
			}
			newPlaidErr = append(newPlaidErr, plaidErr)

			// impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the  transaction details.")
			return UserTransaction{}, nil, newPlaidErr
		}
		transactions := transGetResp.GetTransactions()

		if len(transactions) > 0 {
			totalTransaction = int32(len(transactions))
			userData.ImpartWealthID = impartWealthId
			userData.AccessToken = accessToken
			userinstitution := make(UserInstitutions, 1)
			var transDatawithdateFinalData []Transaction
			var transDataFinalData []TransactionWithDate
			institution := InstitutionToModel(userInstitutionList[0])
			var allDates []string
			for _, act := range transactions {
				if act.AccountId == accountId {
					currentDate := act.Date
					ser.logger.Info(currentDate)
					ser.logger.Info("allDates", zap.Any("allDates", allDates))
					if !checkDateExist(currentDate, allDates) {
						ser.logger.Info("alredy date added", zap.Any("allDates", allDates),
							zap.Any("currentDate", currentDate))

						for _, acnts := range transactions {
							if currentDate == acnts.Date && acnts.AccountId == accountId {
								if !checkDateExist(currentDate, allDates) {
									allDates = append(allDates, currentDate)
								}
								transDatawithdate := TransactionToModel(acnts, userInstitutionList[0].UserInstitutionID)

								transDatawithdateFinalData = append(transDatawithdateFinalData, transDatawithdate)
							}
						}

						if len(transDatawithdateFinalData) > 0 {
							transWIthdate := TransactionWithDate{}
							transWIthdate.Date = currentDate
							transWIthdate.Data = transDatawithdateFinalData
							transDataFinalData = append(transDataFinalData, transWIthdate)
							transDatawithdateFinalData = nil
						} else {
							return UserTransaction{}, nil, newPlaidErr
						}
					}
				}
			}

			if len(transDataFinalData) > 0 {
				userinstitution[0] = institution
				userData.Transactions = transDataFinalData
				userData.TotalTransaction = transGetResp.GetTotalTransactions()

				outOffset := &NextPage{
					Offset: int(gpi.Offset),
				}
				if totalTransaction < gpi.Limit {
					outOffset = nil
				} else {
					outOffset.Offset += int(totalTransaction)
				}
				return userData, outOffset, nil
			}

		}

	} else if userInstitutionList[0].BankType == 2 {
		transInvestGetRequest := plaid.NewInvestmentsTransactionsGetRequest(accessToken, impart.CurrentUTC().AddDate(0, 0, -30).Format("2006-01-02"), impart.CurrentUTC().Format("2006-01-02"))
		data := plaid.NewInvestmentsTransactionsGetRequestOptions()
		var accountIds []string
		accountIds = append(accountIds, accountId)
		transInvestGetRequest.Options = data
		transInvestGetRequest.Options.Count = &gpi.Limit
		transInvestGetRequest.Options.Offset = &gpi.Offset
		transInvestGetRequest.Options.AccountIds = &accountIds
		transGetResp, resp, err := client.PlaidApi.InvestmentsTransactionsGet(ctx).InvestmentsTransactionsGetRequest(
			*transInvestGetRequest,
		).Execute()
		if err != nil || resp.StatusCode == 400 {
			ser.logger.Error("Could not find the user plaid account details.", zap.String("User", impartWealthId),
				zap.String("token", accessToken))
			plaidErr.Msg = "Could not find the  transaction details."
			if resp.StatusCode == 400 {
				defer resp.Body.Close()
				type errorResponse struct {
					ErrorCode string `json:"error_code" `
				}
				newRes := errorResponse{}
				json.NewDecoder(resp.Body).Decode(&newRes)
				if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
					plaidErr.AuthenticationError = true
					plaidErr.Msg = "ITEM_LOGIN_REQUIRED"
				}
				plaidErr.AccessToken = accessToken
			}
			newPlaidErr = append(newPlaidErr, plaidErr)

			// impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the  transaction details.")
			return UserTransaction{}, nil, newPlaidErr
		}
		investTransactions = transGetResp.GetInvestmentTransactions()
		totalTransaction = int32(len(investTransactions))
		if totalTransaction > 0 {
			userData.ImpartWealthID = impartWealthId
			userData.AccessToken = accessToken
			userinstitution := make(UserInstitutions, 1)
			var transDatawithdateFinalData []Transaction
			var transDataFinalData []TransactionWithDate
			institution := InstitutionToModel(userInstitutionList[0])
			var allDates []string
			for _, act := range investTransactions {
				if act.AccountId == accountId {
					currentDate := act.Date
					ser.logger.Info(currentDate)
					ser.logger.Info("allDates", zap.Any("allDates", allDates))
					if !checkDateExist(currentDate, allDates) {
						ser.logger.Info("alredy date added", zap.Any("allDates", allDates),
							zap.Any("currentDate", currentDate))

						for _, acnts := range investTransactions {
							if currentDate == acnts.Date && acnts.AccountId == accountId {
								ser.logger.Info("acnts.Date", zap.Any("acnts.Date", acnts.Date))
								if !checkDateExist(currentDate, allDates) {
									allDates = append(allDates, currentDate)
								}
								ser.logger.Info("aallDates", zap.Any("allDates", allDates))
								transDatawithdate := InvestmentTransactionToModel(acnts, userInstitutionList[0].UserInstitutionID)
								transDatawithdateFinalData = append(transDatawithdateFinalData, transDatawithdate)
							}
						}
						if len(transDatawithdateFinalData) > 0 {
							transWIthdate := TransactionWithDate{}
							transWIthdate.Date = currentDate
							transWIthdate.Data = transDatawithdateFinalData
							transDataFinalData = append(transDataFinalData, transWIthdate)
							transDatawithdateFinalData = nil
						} else {
							return UserTransaction{}, nil, newPlaidErr
						}
					}
				}
			}
			if len(transDataFinalData) > 0 {
				userinstitution[0] = institution
				userData.Transactions = transDataFinalData
				userData.TotalTransaction = transGetResp.GetTotalInvestmentTransactions()
				outOffset := &NextPage{
					Offset: int(gpi.Offset),
				}

				if totalTransaction < gpi.Limit {
					outOffset = nil
				} else {
					outOffset.Offset += int(totalTransaction)
				}
				return userData, outOffset, nil
			}

		}
	}

	return UserTransaction{}, nil, newPlaidErr
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

func InvestmentTransactionToModel(act plaid.InvestmentTransaction, userInstId uint64) Transaction {
	var allDates []string
	if act.Subtype == "" {
		allDates = append(allDates, "investment")
	} else {
		allDates = append(allDates, act.Subtype)
	}
	trans := Transaction{}
	trans.AccountID = act.AccountId
	trans.Amount = act.GetAmount()
	trans.Category = allDates
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
		} else if cfg.Env == config.Development {
			configuration.UseEnvironment(plaid.Sandbox)
		} else {
			configuration.UseEnvironment(plaid.Sandbox)
		}
	}
	client := plaid.NewAPIClient(configuration)
	accountsGetRequest := plaid.NewAccountsGetRequest(accessToken)
	_, response, _ := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*accountsGetRequest,
	).Execute()

	if response.StatusCode == 400 {
		defer response.Body.Close()
		type errorResponse struct {
			ErrorCode string `json:"error_code" `
		}
		newRes := errorResponse{}
		json.NewDecoder(response.Body).Decode(&newRes)
		if newRes.ErrorCode == "ITEM_LOGIN_REQUIRED" {
			return true
		}
	}
	return false
}

func (ser *service) DeletePlaidUserInstitutionAccounts(ctx context.Context, userInstitutionId uint64) impart.Error {
	userInstitutions, err := dbmodels.FindUserInstitution(ctx, ser.db, userInstitutionId)
	if err != nil {
		fmt.Println(err)
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not find the user institution.")
		return impartErr
	}
	_, err = userInstitutions.Delete(ctx, ser.db, false)
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, "Could not delete the record.")
		return impartErr
	}
	return nil
}
