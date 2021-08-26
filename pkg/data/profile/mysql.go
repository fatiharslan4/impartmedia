package profile

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	authdata "github.com/impartwealthapp/backend/pkg/data/auth"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
	"gopkg.in/auth0.v5/management"
)

var _ Store = &mysqlStore{}

type mysqlStore struct {
	logger *zap.Logger
	db     *sql.DB
}

func (m *mysqlStore) GetProfile(ctx context.Context, impartWealthId string) (*dbmodels.Profile, error) {
	out, err := dbmodels.Profiles(dbmodels.ProfileWhere.ImpartWealthID.EQ(impartWealthId)).One(ctx, m.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}
	return out, err
}

func newMysqlStore(db *sql.DB, logger *zap.Logger) *mysqlStore {
	out := &mysqlStore{
		db:     db,
		logger: logger,
	}
	return out
}

func (m *mysqlStore) getUser(ctx context.Context, impartID, authID, email, screenName string) (*dbmodels.User, error) {
	var clause QueryMod
	if impartID != "" {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.ImpartWealthID), impartID)
	} else if authID != "" {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.AuthenticationID), authID)
	} else if email != "" {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.Email), email)
	} else {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.ScreenName), screenName)
	}
	usersWhere := []QueryMod{
		clause,
		Load(dbmodels.UserRels.ImpartWealthProfile),
		Load(dbmodels.UserRels.MemberHiveHives),
		Load(dbmodels.UserRels.ImpartWealthUserDevices),
		Load(dbmodels.UserRels.ImpartWealthUserConfigurations),
		Load(dbmodels.UserRels.ImpartWealthUserAnswers),
	}

	u, err := dbmodels.Users(usersWhere...).One(ctx, m.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, err
}

func (m *mysqlStore) GetUser(ctx context.Context, impartWealthID string) (*dbmodels.User, error) {
	return m.getUser(ctx, impartWealthID, "", "", "")
}

func (m *mysqlStore) GetUserFromAuthId(ctx context.Context, authenticationId string) (*dbmodels.User, error) {
	return m.getUser(ctx, "", authenticationId, "", "")
}

func (m *mysqlStore) GetUserFromEmail(ctx context.Context, email string) (*dbmodels.User, error) {
	return m.getUser(ctx, "", "", email, "")
}

func (m *mysqlStore) GetUserFromScreenName(ctx context.Context, screenName string) (*dbmodels.User, error) {
	return m.getUser(ctx, "", "", "", screenName)
}

func rollbackIfError(tx *sql.Tx, err error, logger *zap.Logger) error {
	rErr := tx.Rollback()
	if rErr != nil {
		logger.Error("unable to rollback transaction", zap.Error(rErr))
		return fmt.Errorf(rErr.Error(), err)
	}
	return err
}

func (m *mysqlStore) CreateUserProfile(ctx context.Context, user *dbmodels.User, profile *dbmodels.Profile) error {
	if user == nil || profile == nil {
		m.logger.Error("user or profile is nil")
		return impart.ErrBadRequest
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	if err = user.Insert(ctx, m.db, boil.Infer()); err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	if err = user.SetImpartWealthProfile(ctx, m.db, true, profile); err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	//add the default hive
	if err = user.SetMemberHiveHives(ctx, m.db, false, &dbmodels.Hive{HiveID: 1}); err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	return tx.Commit()
}

func (m *mysqlStore) UpdateProfile(ctx context.Context, user *dbmodels.User, profile *dbmodels.Profile) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	if user != nil {
		_, err = user.Update(ctx, tx, boil.Infer())
		if err != nil {
			return rollbackIfError(tx, err, m.logger)
		}
	}
	if profile != nil {
		_, err = profile.Update(ctx, tx, boil.Infer())
		if err != nil {
			return rollbackIfError(tx, err, m.logger)
		}
	}
	return tx.Commit()
}

func (m *mysqlStore) DeleteProfile(ctx context.Context, impartWealthID string, hardDelete bool) error {
	u, err := dbmodels.FindUser(ctx, m.db, impartWealthID)
	if err == sql.ErrNoRows || u == nil {
		return impart.ErrNotFound
	}
	if err != nil {
		return err
	}
	_, err = u.Delete(ctx, m.db, hardDelete)
	return err
}

func (m *mysqlStore) GetQuestionnaire(ctx context.Context, name string, version *uint) (*dbmodels.Questionnaire, error) {
	qms := []QueryMod{
		dbmodels.QuestionnaireWhere.Name.EQ(name),
		Load(Rels(dbmodels.QuestionnaireRels.Questions, dbmodels.QuestionRels.Type)),
		Load(Rels(dbmodels.QuestionnaireRels.Questions, dbmodels.QuestionRels.Answers)),
	}

	if version == nil || *version == 0 {
		qms = append(qms, Where(`
				EXISTS (
					select q.name, max(q.version)
                	from questionnaire q
                	where
					q.name = ?
					and q.enabled = true 
					and q.questionnaire_id = questionnaire.questionnaire_id
                	group by q.name
					)`, name))
	} else {
		qms = append(qms, dbmodels.QuestionnaireWhere.Version.EQ(*version))
	}

	questionnaire, err := dbmodels.Questionnaires(qms...).One(ctx, m.db)
	if err != nil {
		m.logger.Error("couldn't fetch questionnaire version",
			zap.String("name", name), zap.Error(err), zap.Error(err))
		return nil, err
	}
	return questionnaire, nil
}

func (m *mysqlStore) GetAllCurrentQuestionnaires(ctx context.Context) (dbmodels.QuestionnaireSlice, error) {
	currentQuestionnaires, err := dbmodels.Questionnaires(
		Where(`EXISTS (select q.name, max(q.version)
                from questionnaire q
                where q.enabled = true and q.questionnaire_id = questionnaire.questionnaire_id
                group by q.name)`),
		Load(Rels(dbmodels.QuestionnaireRels.Questions, dbmodels.QuestionRels.Type)),
		Load(Rels(dbmodels.QuestionnaireRels.Questions, dbmodels.QuestionRels.Answers))).
		All(ctx, m.db)
	if err != nil {
		m.logger.Error("couldn't fetch latest version for questionnaire", zap.Error(err))
		return nil, err
	}

	return currentQuestionnaires, nil
}

func (m *mysqlStore) SaveUserQuestionnaire(ctx context.Context, answers dbmodels.UserAnswerSlice) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer impart.CommitRollbackLogger(tx, err, m.logger)
	for _, a := range answers {
		err := a.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}

		var userDemo dbmodels.UserDemographic
		err = dbmodels.NewQuery(
			qm.Select("*"),
			qm.Where("answer_id = ?", a.AnswerID),
			qm.From("user_demographic"),
		).Bind(ctx, m.db, &userDemo)

		if err == nil {
			existData := &userDemo
			existData.UserCount = existData.UserCount + 1
			_, err = existData.Update(ctx, m.db, boil.Infer())
		}

		var hiveUserdemo dbmodels.HiveUserDemographic
		err = dbmodels.NewQuery(
			qm.Select("*"),
			qm.Where("answer_id = ?", a.AnswerID),
			qm.Where("hive_id = ?", DefaultHiveId),
			qm.From("hive_user_demographic"),
		).Bind(ctx, m.db, &hiveUserdemo)

		if err == nil {
			existUserData := &hiveUserdemo
			existUserData.UserCount = existUserData.UserCount + 1
			_, err = existUserData.Update(ctx, m.db, boil.Infer())
		}
	}
	tx.Commit()
	return nil
}

func (m *mysqlStore) GetUserQuestionnaires(ctx context.Context, impartWealthId string, questionnaireName *string) (dbmodels.QuestionnaireSlice, error) {
	qm := []QueryMod{
		dbmodels.UserAnswerWhere.ImpartWealthID.EQ(impartWealthId),
	}

	qm = append(qm, Load(Rels(dbmodels.UserAnswerRels.Answer, dbmodels.AnswerRels.Question, dbmodels.QuestionRels.Questionnaire)))
	qm = append(qm, Load(Rels(dbmodels.UserAnswerRels.Answer, dbmodels.AnswerRels.Question, dbmodels.QuestionRels.Type)))

	userAnswers, err := dbmodels.UserAnswers(qm...).All(ctx, m.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbmodels.QuestionnaireSlice{}, impart.ErrNotFound
		}
	}

	dedupMap := make(map[uint]*dbmodels.Questionnaire)
	for _, a := range userAnswers {
		q, ok := dedupMap[a.R.Answer.R.Question.R.Questionnaire.QuestionnaireID]
		if !ok {
			q = a.R.Answer.R.Question.R.Questionnaire
			dedupMap[q.QuestionnaireID] = q
		}
		if questionnaireName != nil && q.Name == *questionnaireName {

		}
	}

	//Build the output list, if we're filtering by name only include those, otherwise include all
	out := make(dbmodels.QuestionnaireSlice, 0)
	for _, v := range dedupMap {
		if questionnaireName == nil || v.Name == *questionnaireName {
			out = append(out, v)
		}
	}

	if len(out) == 0 {
		return dbmodels.QuestionnaireSlice{}, impart.ErrNotFound
	}

	return out, nil
}

//  GetUserDevice : Get the user device
func (m *mysqlStore) GetUserDevice(ctx context.Context, token string, impartID string, deviceToken string) (*dbmodels.UserDevice, error) {
	where := []QueryMod{}
	if impartID != "" {
		where = append(where, Where(fmt.Sprintf("%s = ?", dbmodels.UserDeviceColumns.ImpartWealthID), impartID))
	}
	if token != "" {
		where = append(where, Where(fmt.Sprintf("%s = ?", dbmodels.UserDeviceColumns.Token), token))
	}
	if deviceToken != "" {
		if deviceToken == "__NILL__" {
			where = append(where, Where(fmt.Sprintf("%s = ?", dbmodels.UserDeviceColumns.DeviceToken), ""))
		} else {
			where = append(where, Where(fmt.Sprintf("%s = ?", dbmodels.UserDeviceColumns.DeviceToken), deviceToken))
		}
	}

	where = append(where, Load(dbmodels.UserDeviceRels.ImpartWealth))
	where = append(where, Load(dbmodels.UserDeviceRels.NotificationDeviceMappings))

	device, err := dbmodels.UserDevices(where...).One(ctx, m.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return device, err
}

// CreateUserDevice
func (m *mysqlStore) CreateUserDevice(ctx context.Context, device *dbmodels.UserDevice) (*dbmodels.UserDevice, error) {
	if device == nil {
		m.logger.Error("device is nil")
		return nil, impart.ErrBadRequest
	}
	uuid := uuid.New()
	device.Token = uuid.String()

	err := device.Insert(ctx, m.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	return m.GetUserDevice(ctx, device.Token, "", "")
}

// AddUserConfigurations
func (m *mysqlStore) CreateUserConfigurations(ctx context.Context, conf *dbmodels.UserConfiguration) (*dbmodels.UserConfiguration, error) {
	if conf.ImpartWealthID == "" {
		m.logger.Error("impartWealthID is nil")
		return nil, impart.ErrBadRequest
	}
	err := conf.Insert(ctx, m.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// Edit User Configurations
func (m *mysqlStore) EditUserConfigurations(ctx context.Context, conf *dbmodels.UserConfiguration) (*dbmodels.UserConfiguration, error) {
	if conf.ImpartWealthID == "" {
		m.logger.Error("impartWealthID is nil")
		return nil, impart.ErrBadRequest
	}
	if _, err := conf.Update(ctx, m.db, boil.Infer()); err != nil {
		return nil, err
	}
	return conf, conf.Reload(ctx, m.db)
}

// GetUserConfigurations
func (m *mysqlStore) GetUserConfigurations(ctx context.Context, impartWealthID string) (*dbmodels.UserConfiguration, error) {
	if impartWealthID == "" {
		m.logger.Error("impartWealthID is nil")
		return nil, impart.ErrBadRequest
	}
	where := []QueryMod{
		Where(fmt.Sprintf("%s = ?", dbmodels.UserConfigurationColumns.ImpartWealthID), impartWealthID),
		Load(dbmodels.UserConfigurationRels.ImpartWealth),
	}

	configurations, err := dbmodels.UserConfigurations(where...).One(ctx, m.db)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return configurations, nil
}

// GetUserNotificationMappData
func (m *mysqlStore) GetUserNotificationMappData(input models.MapArgumentInput) (*dbmodels.NotificationDeviceMapping, error) {
	where := []QueryMod{}
	if input.ImpartWealthID != "" {
		where = append(where, dbmodels.NotificationDeviceMappingWhere.ImpartWealthID.EQ(input.ImpartWealthID))
	}
	if input.DeviceToken != "" {
		// where = append(where, dbmodels.NotificationDeviceMappingWhere.UserDeviceID.EQ(input.DeviceToken))
	}
	if input.DeviceToken != "" {
		where = append(where, qm.InnerJoin("user_devices ON user_devices.token = notification_device_mapping.user_device_id"))
		where = append(where, qm.Where("user_devices.device_token=?", input.DeviceToken))
	}

	mapData, err := dbmodels.NotificationDeviceMappings(where...).One(input.Ctx, m.db)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return mapData, nil
}

//  DeleteUserNotificationMappData
//  Delete the user notification map details
func (m *mysqlStore) DeleteUserNotificationMappData(input models.MapArgumentInput) error {
	where := []QueryMod{}
	if input.ImpartWealthID != "" {
		where = append(where, dbmodels.NotificationDeviceMappingWhere.ImpartWealthID.EQ(input.ImpartWealthID))
	}
	if input.DeviceToken != "" {
		// where = append(where, dbmodels.NotificationDeviceMappingWhere.UserDeviceID.EQ(input.DeviceToken))
	}
	if input.DeviceToken != "" {
		where = append(where, qm.Where("user_device_id IN (select token from user_devices where device_token = ?)", input.DeviceToken))
	}

	_, err := dbmodels.NotificationDeviceMappings(where...).DeleteAll(input.Ctx, m.db)
	if err != nil {
		return err
	}
	if err == sql.ErrNoRows {
		return impart.ErrNotFound
	}

	return nil
}

// DeleteUserNotificationMappData
// Delete the user notification map details
func (m *mysqlStore) UpdateExistingNotificationMappData(input models.MapArgumentInput, notifyStatus bool) error {
	where := []QueryMod{}
	// where impart id provided and negate is false
	if input.ImpartWealthID != "" && !input.Negate {
		where = append(where, dbmodels.NotificationDeviceMappingWhere.ImpartWealthID.EQ(input.ImpartWealthID))
	}
	// where impart id provided and required negate
	if input.ImpartWealthID != "" && input.Negate {
		where = append(where, dbmodels.NotificationDeviceMappingWhere.ImpartWealthID.NEQ(input.ImpartWealthID))
	}
	if input.Token != "" {
		where = append(where, dbmodels.NotificationDeviceMappingWhere.UserDeviceID.EQ(input.Token))
	}
	if input.DeviceToken != "" {
		where = append(where, qm.Where("user_device_id IN (select token from user_devices where device_token = ?)", input.DeviceToken))
	}
	if input.DeviceID != "" {
		where = append(where, qm.Where("user_device_id IN (select token from user_devices where device_id = ?)", input.DeviceID))
	}
	_, err := dbmodels.NotificationDeviceMappings(where...).UpdateAll(input.Ctx, m.db, dbmodels.M{
		"notify_status": notifyStatus,
	})
	if err != nil {
		return err
	}
	if err == sql.ErrNoRows {
		return impart.ErrNotFound
	}

	return nil
}

// CreateUserNotificationMappData
// create user notificatoin map data
func (m *mysqlStore) CreateUserNotificationMappData(ctx context.Context, data *dbmodels.NotificationDeviceMapping) (*dbmodels.NotificationDeviceMapping, error) {
	if data == nil {
		m.logger.Error("maping data is nil")
		return nil, impart.ErrBadRequest
	}

	err := data.Insert(ctx, m.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Block a user
func (m *mysqlStore) BlockUser(ctx context.Context, user *dbmodels.User, status bool) error {
	// set the blocked status
	user.Blocked = status
	_, err := user.Update(ctx, m.db, boil.Infer())
	if err != nil {
		m.logger.Error("unable to block user", zap.Any("error", err))
		return fmt.Errorf("unable to block")
	}
	return nil
}

func (m *mysqlStore) UpdateDeviceToken(ctx context.Context, device *dbmodels.UserDevice, deviceToken string) error {
	device.DeviceToken = deviceToken
	_, err := device.Update(ctx, m.db, boil.Infer())
	if err != nil {
		m.logger.Error("unable to update device token user", zap.Any("error", err))
		return fmt.Errorf("unable to update device token")
	}
	return nil
}
func (m *mysqlStore) UpdateDevice(ctx context.Context, device *dbmodels.UserDevice) error {
	_, err := device.Update(ctx, m.db, boil.Infer())
	if err != nil {
		m.logger.Error("unable to update device", zap.Any("error", err))
		return fmt.Errorf("unable to update")
	}
	return nil
}

func (m *mysqlStore) DeleteExceptUserDevice(ctx context.Context, impartID string, deviceToken string, refToken string) error {
	// Delete a slice of pilots from the database
	_, err := dbmodels.UserDevices(
		dbmodels.UserDeviceWhere.ImpartWealthID.EQ(impartID),
		dbmodels.UserDeviceWhere.DeviceToken.EQ(deviceToken),
		dbmodels.UserDeviceWhere.Token.NEQ(refToken)).DeleteAll(ctx, m.db, true)

	if err != nil {
		return fmt.Errorf("error occured during delete non wanted devices %v", err)
	}

	return nil
}

func (m *mysqlStore) UpdateUserDemographic(ctx context.Context, answerIds []uint, status bool) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	for _, a := range answerIds {
		var userDemo dbmodels.UserDemographic
		err = dbmodels.NewQuery(
			qm.Select("*"),
			qm.Where("answer_id = ?", a),
			qm.From("user_demographic"),
		).Bind(ctx, m.db, &userDemo)

		if err == nil {
			existData := &userDemo
			if status {
				existData.UserCount = existData.UserCount + 1
			} else {
				existData.UserCount = existData.UserCount - 1
			}
			_, err = existData.Update(ctx, m.db, boil.Infer())
			if err != nil {
				return rollbackIfError(tx, err, m.logger)
			}
		}
	}
	return tx.Commit()
}

func (m *mysqlStore) GetMakeUp(ctx context.Context) (interface{}, error) {

	dataMap := make(map[int]map[string]interface{})
	userAnswers, err := dbmodels.UserDemographics(
		Load(Rels(dbmodels.UserDemographicRels.Answer, dbmodels.AnswerRels.Question, dbmodels.QuestionRels.Questionnaire)),
		Load(Rels(dbmodels.UserDemographicRels.Answer, dbmodels.AnswerRels.Question, dbmodels.QuestionRels.Type)),
	).All(ctx, m.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return dataMap, impart.ErrNotFound
		}
	}
	dedupMap := make(map[uint]*dbmodels.Questionnaire)

	totalCnt := 0
	percentageTotal := 0.0

	if len(userAnswers) == 0 {
		return dataMap, impart.ErrNotFound
	}

	indexes := make(map[uint]int)

	for _, a := range userAnswers {
		q, ok := dedupMap[a.R.Answer.R.Question.R.Questionnaire.QuestionnaireID]
		if !ok {
			q = a.R.Answer.R.Question.R.Questionnaire
			dedupMap[q.QuestionnaireID] = q
		}

		qIDInt := int(q.QuestionnaireID)
		questionIDstr := strconv.Itoa(int(a.R.Answer.R.Question.QuestionID))
		answerIDstr := strconv.Itoa(int(a.R.Answer.AnswerID))

		// if the index not exists
		if _, ok := dataMap[qIDInt]; !ok {
			dataMap[qIDInt] = make(map[string]interface{})
		}

		// check questions index exists
		if _, ok := dataMap[qIDInt][questionIDstr]; !ok {
			totalCnt = 0
			percentageTotal = 0.0
			for _, ans := range userAnswers {
				if ans.R.Answer.R.Question.QuestionID == uint(a.R.Answer.R.Question.QuestionID) {
					totalCnt = totalCnt + ans.UserCount
				}
			}
			indexes[uint(a.R.Answer.R.Question.QuestionID)] = totalCnt
			dataMap[qIDInt][questionIDstr] = make(map[string]interface{})
			dataMap[qIDInt][questionIDstr].(map[string]interface{})["questions"] = make(map[string]interface{})
		}

		// check answers index exists in
		if _, ok := dataMap[qIDInt][questionIDstr].(map[string]interface{})["questions"].(map[string]interface{})[answerIDstr]; !ok {
			dataMap[qIDInt][questionIDstr].(map[string]interface{})["questions"].(map[string]interface{})[answerIDstr] = make(map[string]interface{})
		}

		// set the array data
		dataMap[qIDInt][questionIDstr].(map[string]interface{})["name"] = a.R.Answer.R.Question.QuestionName
		dataMap[qIDInt][questionIDstr].(map[string]interface{})["questionText"] = a.R.Answer.R.Question.Text
		percentage := 0.0
		if a.UserCount > 0 {
			percentage = float64(a.UserCount) / float64(indexes[uint(a.R.Answer.R.Question.QuestionID)]) * 100
		}

		per, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", percentage), 64)
		percentageTotal = percentageTotal + per
		if percentageTotal > 100 {
			dif := percentageTotal - 100
			per = per - dif
		}
		dataMap[qIDInt][questionIDstr].(map[string]interface{})["questions"].(map[string]interface{})[answerIDstr] = map[string]string{
			"id":    strconv.Itoa(int(a.R.Answer.AnswerID)),
			"title": a.R.Answer.AnswerName,
			"text":  a.R.Answer.Text,
			"count": strconv.Itoa(a.UserCount),
			// "percentage": fmt.Sprintf("%f", percentage),
			"percentage": fmt.Sprintf("%.1f", per),
		}
	}

	return dataMap, nil
}

func (m *mysqlStore) DeleteUserProfile(ctx context.Context, gpi models.DeleteUserInput, hardDelete bool) impart.Error {
	userToDelete, err := m.GetUser(ctx, gpi.ImpartWealthID)
	if err != nil {
		return impart.NewError(err, fmt.Sprintf("couldn't find profile for impartWealthID %s", gpi.ImpartWealthID))
	}
	hiveid := DefaultHiveId
	for _, h := range userToDelete.R.MemberHiveHives {
		hiveid = h.HiveID
	}
	if hardDelete {
		err = m.DeleteProfile(ctx, gpi.ImpartWealthID, hardDelete)
		if err != nil {
			return impart.NewError(err, "unable to retrieve profile")
		}
		return nil
	}
	existingDBProfile := userToDelete.R.ImpartWealthProfile
	exitingUserAnswer := userToDelete.R.ImpartWealthUserAnswers
	answerIds := make([]uint, len(exitingUserAnswer))
	for i, a := range exitingUserAnswer {
		answerIds[i] = a.AnswerID
	}
	userEmail := userToDelete.Email
	screenName := userToDelete.ScreenName
	userToDelete = models.UpdateToUserDB(userToDelete, gpi, true, screenName, userEmail)
	err = m.UpdateProfile(ctx, userToDelete, existingDBProfile)
	if err != nil {
		m.logger.Error("Delete user requset failed", zap.String("deleteUser", userToDelete.ImpartWealthID),
			zap.String("contextUser", userToDelete.ImpartWealthID))

		return impart.NewError(err, "User Deletion failed")

	}
	if !userToDelete.Blocked {
		err = m.UpdateUserDemographic(ctx, answerIds, false)
		err = m.UpdateHiveUserDemographic(ctx, answerIds, false, hiveid)
	}

	mngmnt, err := authdata.NewImpartManagementClient()
	if err != nil {
		////revert the server update
		userToDelete = models.UpdateToUserDB(userToDelete, gpi, false, screenName, userEmail)

		err = m.UpdateProfile(ctx, userToDelete, existingDBProfile)
		if err != nil {
			m.logger.Error("Delete user requset failed in auth 0 then revert the server", zap.String("deleteUser", userToDelete.ImpartWealthID),
				zap.String("contextUser", userToDelete.ImpartWealthID))
		}
		if !userToDelete.Blocked {
			err = m.UpdateUserDemographic(ctx, answerIds, true)
			if err != nil {
				m.logger.Error("Delete user requset failed in auth 0 then revert the server- user demographic falied.", zap.String("deleteUser", userToDelete.ImpartWealthID),
					zap.String("contextUser", userToDelete.ImpartWealthID))
			}
			err = m.UpdateHiveUserDemographic(ctx, answerIds, true, hiveid)
		}
		return impart.NewError(err, "User Deletion failed")

	}
	userEmail = fmt.Sprintf("%s%s", userToDelete.ImpartWealthID, userEmail)
	userUp := management.User{
		Email: &userEmail,
	}

	errDel := mngmnt.User.Update(*&userToDelete.AuthenticationID, &userUp)
	if errDel != nil {

		//revert the server update
		userToDelete = models.UpdateToUserDB(userToDelete, gpi, true, screenName, userEmail)

		err = m.UpdateProfile(ctx, userToDelete, existingDBProfile)
		if err != nil {
			m.logger.Error("Delete user requset failed in auth 0 then revert the server- user failed.", zap.String("deleteUser", userToDelete.ImpartWealthID),
				zap.String("contextUser", userToDelete.ImpartWealthID))
		}
		if !userToDelete.Blocked {
			err = m.UpdateUserDemographic(ctx, answerIds, true)
			if err != nil {
				m.logger.Error("Delete user requset failed in auth 0 then revert the server- user demographic falied.", zap.String("deleteUser", userToDelete.ImpartWealthID),
					zap.String("contextUser", userToDelete.ImpartWealthID))
			}
			err = m.UpdateHiveUserDemographic(ctx, answerIds, true, hiveid)
		}
		return impart.NewError(err, "User Deletion failed")
	}
	return nil
}

func (m *mysqlStore) UpdateHiveUserDemographic(ctx context.Context, answerIds []uint, status bool, hiveId uint64) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	for _, a := range answerIds {
		var userDemo dbmodels.HiveUserDemographic
		err = dbmodels.NewQuery(
			qm.Select("*"),
			qm.Where("answer_id = ?", a),
			qm.Where("hive_id = ?", int(hiveId)),
			qm.From("hive_user_demographic"),
		).Bind(ctx, m.db, &userDemo)
		if err != nil {
		}
		if err == nil {
			existData := &userDemo
			if status {
				existData.UserCount = existData.UserCount + 1
			} else {
				existData.UserCount = existData.UserCount - 1
			}
			_, err = existData.Update(ctx, m.db, boil.Infer())
			if err != nil {
				return rollbackIfError(tx, err, m.logger)
			}
		}
	}
	return tx.Commit()
}
