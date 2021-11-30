package profile

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/beeker1121/mailchimp-go/lists/members"
	"github.com/google/uuid"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	authdata "github.com/impartwealthapp/backend/pkg/data/auth"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
	"gopkg.in/auth0.v5/management"
)

var _ Store = &mysqlStore{}

type mysqlStore struct {
	logger              *zap.Logger
	db                  *sql.DB
	notificationService impart.NotificationService
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

func newMysqlStore(db *sql.DB, logger *zap.Logger, notificationService impart.NotificationService) *mysqlStore {
	out := &mysqlStore{
		db:                  db,
		logger:              logger,
		notificationService: notificationService,
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

		// var userDemo dbmodels.UserDemographic
		// err = dbmodels.NewQuery(
		// 	qm.Select("*"),
		// 	qm.Where("answer_id = ?", a.AnswerID),
		// 	qm.From("user_demographic"),
		// ).Bind(ctx, m.db, &userDemo)

		// if err == nil {
		// 	existData := &userDemo
		// 	existData.UserCount = existData.UserCount + 1
		// 	_, err = existData.Update(ctx, m.db, boil.Infer())
		// }

		// var hiveUserdemo dbmodels.HiveUserDemographic
		// err = dbmodels.NewQuery(
		// 	qm.Select("*"),
		// 	qm.Where("answer_id = ?", a.AnswerID),
		// 	qm.Where("hive_id = ?", DefaultHiveId),
		// 	qm.From("hive_user_demographic"),
		// ).Bind(ctx, m.db, &hiveUserdemo)

		// if err == nil {
		// 	existUserData := &hiveUserdemo
		// 	existUserData.UserCount = existUserData.UserCount + 1
		// 	_, err = existUserData.Update(ctx, m.db, boil.Infer())
		// }
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
	// hiveid := DefaultHiveId
	// for _, h := range userToDelete.R.MemberHiveHives {
	// 	hiveid = h.HiveID
	// }
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
	orgEmail := userToDelete.Email
	screenName := userToDelete.ScreenName
	userToDelete = models.UpdateToUserDB(userToDelete, gpi, true, screenName, userEmail)

	err = m.UpdateProfile(ctx, userToDelete, existingDBProfile)
	if err != nil {
		m.logger.Error("Delete user requset failed", zap.String("deleteUser", userToDelete.ImpartWealthID),
			zap.String("contextUser", userToDelete.ImpartWealthID))

		return impart.NewError(err, "User Deletion failed")

	}
	currTime := time.Now().In(boil.GetLocation())
	golangDateTime := currTime.Format("2006-01-02 15:04:05.000")

	if userToDelete.Admin {
		postDeleteQuery := fmt.Sprintf(`
		update post
		join ( select comment.post_id,count(comment_id) as count, comment.impart_wealth_id
			from comment
			join post on post.post_id=comment.post_id
			where post.deleted_at is null
			and comment.deleted_at is null
			and comment.impart_wealth_id = '%s'
			group by comment.post_id,comment.impart_wealth_id)
		post_comment
		on post_comment.post_id=post.post_id
		set comment_count= comment_count-post_comment.count
		where comment_count>=post_comment.count;

		update comment 
		set deleted_at ='%s'
		where impart_wealth_id = '%s'
		and deleted_at is null;

	Update  hive
                join post on post.post_id=hive.pinned_post_id
                set pinned_post_id=null
                where   pinned_post_id in ( select post_id from post
                where impart_wealth_id  = '%s');

	update post
	set deleted_at='%s',pinned=false
	where impart_wealth_id = '%s'
	and deleted_at is null;`,
			userToDelete.ImpartWealthID, golangDateTime, userToDelete.ImpartWealthID, userToDelete.ImpartWealthID, golangDateTime, userToDelete.ImpartWealthID)

		_, err = queries.Raw(postDeleteQuery).ExecContext(ctx, m.db)

		m.logger.Info(postDeleteQuery)
		if err != nil {
			m.logger.Error("query failed", zap.Any("query", err), zap.Any("postDeleteQuery", postDeleteQuery))
		}
	}

	// var waitGrp sync.WaitGroup
	if userToDelete.R.MemberHiveHives != nil {
		if userToDelete.R.MemberHiveHives[0].NotificationTopicArn.String != "" {
			// waitGrp.Add(1)
			go func() {
				// defer waitGrp.Done()
				err := m.notificationService.UnsubscribeTopicForAllDevice(ctx, userToDelete.ImpartWealthID, userToDelete.R.MemberHiveHives[0].NotificationTopicArn.String)
				if err != nil {
					m.logger.Error("SubscribeTopic", zap.String("DeviceToken", userToDelete.R.MemberHiveHives[0].NotificationTopicArn.String),
						zap.Error(err))
				}
			}()
			// waitGrp.Wait()
		}
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
		return impart.NewError(err, "User Deletion failed")
	}
	// delete user from mailChimp
	cfg, _ := config.GetImpart()
	// waitGrp.Add(2)
	go func() {
		// defer waitGrp.Done()
		err = members.Delete(cfg.MailchimpAudienceId, orgEmail)
		if err != nil {
			m.logger.Error("Delete user requset failed in MailChimp", zap.String("deleteUser", userToDelete.ImpartWealthID),
				zap.String("contextUser", userToDelete.ImpartWealthID))
		}
	}()
	// go impart.UserDemographicsUpdate(&waitGrp, ctx, m.db, true, true)
	go impart.UserDemographicsUpdate(ctx, m.db, true, true)
	// waitGrp.Wait()
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
func (m *mysqlStore) getUserAll(ctx context.Context, impartWealthids []interface{}, isUserDelete bool, includeUsers int, excludeHive uint64, dates []time.Time) (dbmodels.UserSlice, error) {
	// var clause QueryMod
	// if impartWealthids != nil {
	// 	// clause = WhereIn(fmt.Sprintf("%s in ?", dbmodels.UserColumns.ImpartWealthID), impartWealthids...)
	// }
	newcluse := Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.Blocked), false)
	usersWhere := []QueryMod{
		// clause,
		newcluse,
		Load(dbmodels.UserRels.ImpartWealthProfile),
		Load(dbmodels.UserRels.MemberHiveHives),
		Load(dbmodels.UserRels.ImpartWealthUserDevices),
		Load(dbmodels.UserRels.ImpartWealthUserConfigurations),
		Load(dbmodels.UserRels.ImpartWealthUserAnswers),
	}
	if impartWealthids != nil {
		usersWhere = append(usersWhere, WhereIn(fmt.Sprintf("%s in ?", dbmodels.UserColumns.ImpartWealthID), impartWealthids...))
	}
	if isUserDelete {
		usersWhere = append(usersWhere, Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.SuperAdmin), false))
	}
	if includeUsers == impart.ExcludeAdmin {
		usersWhere = append(usersWhere, Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.Admin), false))
	} else if includeUsers == impart.IncludeAdmin {
		usersWhere = append(usersWhere, Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.Admin), true))
	}
	if excludeHive > 0 {
		usersWhere = append(usersWhere, InnerJoin("`hive_members` on `user`.`impart_wealth_id` = `hive_members`.`member_impart_wealth_id`"),
			Where("`hive_members`.`member_hive_id`!=?", excludeHive))
	}

	if dates != nil {
		usersWhere = append(usersWhere, Where(fmt.Sprintf("%s Between ? and ?", dbmodels.UserColumns.HiveUpdatedAt), dates[1], dates[0]))
	}

	u, err := dbmodels.Users(usersWhere...).All(ctx, m.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, err
}

func (m *mysqlStore) DeleteBulkUserProfile(ctx context.Context, userDetails dbmodels.UserSlice, hardDelete bool) error {
	adminImpartWealthIds := ""
	impartWealthIds := ""
	currTime := time.Now().In(boil.GetLocation())
	golangDateTime := currTime.Format("2006-01-02 15:04:05.000")
	mngmnt, err := authdata.NewImpartManagementClient()
	if err != nil {
		return err
	}
	for _, user := range userDetails {
		impartWealthIds = fmt.Sprintf("%s '%s' ,", impartWealthIds, user.ImpartWealthID)
		if user.Admin {
			adminImpartWealthIds = fmt.Sprintf("%s '%s' ,", adminImpartWealthIds, user.ImpartWealthID)
		}
		// var waitGrp sync.WaitGroup
		if user.R.MemberHiveHives != nil {
			if user.R.MemberHiveHives[0].NotificationTopicArn.String != "" {
				// waitGrp.Add(1)
				go func() {
					// defer waitGrp.Done()
					err := m.notificationService.UnsubscribeTopicForAllDevice(ctx, user.ImpartWealthID, user.R.MemberHiveHives[0].NotificationTopicArn.String)
					if err != nil {
						m.logger.Error("SubscribeTopic", zap.String("DeviceToken", user.R.MemberHiveHives[0].NotificationTopicArn.String),
							zap.Error(err))
					}
				}()
				// waitGrp.Wait()
			}
		}
	}
	query := ""
	if impartWealthIds != "" {
		impartWealthIds = strings.Trim(impartWealthIds, ",")
		updateUser := fmt.Sprintf(`Update user 
		set deleted_at='%s' ,
		 email=CONCAT(email, "-", impart_wealth_id),
		 screen_name=CONCAT(screen_name, "-", impart_wealth_id),
		 deleted_by_admin=true 
		 where impart_wealth_id in(%s);
		`, golangDateTime, impartWealthIds)

		query = updateUser
	}
	if adminImpartWealthIds != "" {
		adminImpartWealthIds = strings.Trim(adminImpartWealthIds, ",")
		postDeleteQuery := fmt.Sprintf(`
		update post
			join ( select comment.post_id,count(comment_id) as count, comment.impart_wealth_id
				from comment
				join post on post.post_id=comment.post_id
				where post.deleted_at is null
				and comment.deleted_at is null
				and comment.impart_wealth_id in (%s) 
				group by comment.post_id,comment.impart_wealth_id)
			post_comment
			on post_comment.post_id=post.post_id
			set comment_count= comment_count-post_comment.count
			where comment_count>=post_comment.count;

		update comment 
		set deleted_at ='%s'
		where impart_wealth_id in (%s) 
		and deleted_at is null;

		Update  hive
                join post on post.post_id=hive.pinned_post_id
                set pinned_post_id=null
                where   pinned_post_id in ( select post_id from post
                where impart_wealth_id in ( %s ));

		update post
		set deleted_at='%s',pinned=false
		where impart_wealth_id in (%s)
		and deleted_at is null;`,
			adminImpartWealthIds, golangDateTime, adminImpartWealthIds, adminImpartWealthIds, golangDateTime, adminImpartWealthIds)

		query = fmt.Sprintf("%s %s", query, postDeleteQuery)
	}
	_, err = queries.Raw(query).ExecContext(ctx, m.db)
	m.logger.Info(query)
	if err != nil {
		m.logger.Error("query failed", zap.Any("query", err))
		return err
	}
	// var waitGrpMailChimp sync.WaitGroup
	// waitGrpMailChimp.Add(len(userDetails))
	go func() {
		for _, user := range userDetails {
			// defer waitGrpMailChimp.Done()
			email := fmt.Sprintf("%s-%s", user.ImpartWealthID, user.Email)
			userUp := management.User{
				Email: &email,
			}
			err = mngmnt.User.Update(*&user.AuthenticationID, &userUp)
			if err != nil {
				m.logger.Error("Auth update failed", zap.Any("user.Email", user.Email), zap.Any("query", err))
			}
		}
	}()
	// waitGrpMailChimp.Add(1)
	go impart.UserDemographicsUpdate(ctx, m.db, true, true)
	// waitGrpMailChimp.Wait()
	return nil
}

func (m *mysqlStore) UpdateBulkUserProfile(ctx context.Context, userDetails dbmodels.UserSlice, hardDelete bool, userUpdate *models.UserUpdate) (*models.UserUpdate, error) {
	updateQuery := ""
	// updateHiveDemographic := ""
	impartWealthIds := ""
	updateHivequery := ""
	existinghiveid := DefaultHiveId
	// userHiveDemoexist := make(map[uint64]map[uint64]int)
	var existingHive *dbmodels.Hive
	var newHive *dbmodels.Hive
	if userUpdate.Type == impart.AddToHive {
		var err error
		newHive, err = dbmodels.FindHive(ctx, m.db, userUpdate.HiveID)
		if err != nil {
			return userUpdate, err
		}
	}
	// dbhiveUserDemographic, err := dbmodels.HiveUserDemographics().All(ctx, m.db)
	// for _, p := range dbhiveUserDemographic {
	// 	data := userHiveDemoexist[uint64(p.HiveID)]
	// 	if data == nil {
	// 		count := make(map[uint64]int)
	// 		count[uint64(p.AnswerID)] = int(p.UserCount)
	// 		userHiveDemoexist[uint64(p.HiveID)] = count
	// 	} else {
	// 		data[uint64(p.AnswerID)] = int(p.UserCount)
	// 	}
	// }
	for _, user := range userDetails {
		userUpdateposition := 0
		for i := range userUpdate.Users {
			if userUpdate.Users[i].ImpartWealthID == user.ImpartWealthID {
				userUpdateposition = i
				break
			}
		}
		if userUpdate.Type == impart.AddToAdmin {
			if user.Admin {
				userUpdate.Users[userUpdateposition].Message = "User is already admin."
			} else {
				rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
				admin := impart.GetAvatharLettersAdmin()
				adminindex := rand.Intn(len(admin))
				adminColor := admin[adminindex]

				query := fmt.Sprintf("Update user set admin=true , avatar_background='%s'   where impart_wealth_id='%s';", adminColor, user.ImpartWealthID)
				updateQuery = fmt.Sprintf("%s %s", updateQuery, query)

				userUpdate.Users[userUpdateposition].Value = 1

				if user.R.MemberHiveHives != nil {
					if user.R.MemberHiveHives[0].NotificationTopicArn.String != "" {
						err := m.notificationService.UnsubscribeTopicForAllDevice(ctx, user.ImpartWealthID, user.R.MemberHiveHives[0].NotificationTopicArn.String)
						if err != nil {
							m.logger.Error("SubscribeTopic", zap.String("DeviceToken", user.R.MemberHiveHives[0].NotificationTopicArn.String),
								zap.Error(err))
						}
					}
				}
			}
		} else if userUpdate.Type == impart.AddToWaitlist {
			for _, h := range user.R.MemberHiveHives {
				existinghiveid = h.HiveID
			}
			existingHive, _ = dbmodels.FindHive(ctx, m.db, existinghiveid)
			if existinghiveid == DefaultHiveId {
				userUpdate.Users[userUpdateposition].Message = "User is already on waitlist."
			} else {
				query := fmt.Sprintf("delete from `hive_members` where `member_impart_wealth_id` ='%s'; insert into `hive_members` (`member_impart_wealth_id`, `member_hive_id`) values ('%s', %d);", user.ImpartWealthID, user.ImpartWealthID, DefaultHiveId)
				updateHivequery = fmt.Sprintf("%s %s", updateHivequery, query)
				// exitingUserAnswer := user.R.ImpartWealthUserAnswers

				// if !user.Blocked {
				// 	for _, answer := range exitingUserAnswer {
				// 		if userHiveDemoexist[existinghiveid][uint64(answer.AnswerID)] > 0 {
				// 			userHiveDemoexist[existinghiveid][uint64(answer.AnswerID)] = userHiveDemoexist[existinghiveid][uint64(answer.AnswerID)] - 1
				// 		}
				// 		userHiveDemoexist[DefaultHiveId][uint64(answer.AnswerID)] = userHiveDemoexist[DefaultHiveId][uint64(answer.AnswerID)] + 1
				// 	}
				// }
				userUpdate.Users[userUpdateposition].Value = 1

				if existingHive != nil {
					if existingHive.NotificationTopicArn.String != "" {
						err := m.notificationService.UnsubscribeTopicForAllDevice(ctx, user.ImpartWealthID, existingHive.NotificationTopicArn.String)
						if err != nil {
							m.logger.Error("SubscribeTopic", zap.String("DeviceToken", existingHive.NotificationTopicArn.String),
								zap.Error(err))
						}
					}
				}
			}
		} else if userUpdate.Type == impart.AddToHive {
			m.logger.Info("addtohive started", zap.String("query", impart.AddToHive))
			m.logger.Info("user", zap.String("query", user.ImpartWealthID))
			for _, h := range user.R.MemberHiveHives {
				existinghiveid = h.HiveID
				// existingHive = h
			}
			existingHive, _ = dbmodels.FindHive(ctx, m.db, existinghiveid)
			m.logger.Info("addtohive started- existing hive", zap.Any("existingHive", existingHive))
			m.logger.Info("user-hive", zap.String("query", fmt.Sprintf("%d", existinghiveid)))
			if existinghiveid == userUpdate.HiveID {
				userUpdate.Users[userUpdateposition].Message = "User is already on hive."
			} else {
				query := fmt.Sprintf("delete from hive_members where member_impart_wealth_id = '%s'; insert into hive_members (member_impart_wealth_id, member_hive_id) values ('%s', %d);", user.ImpartWealthID, user.ImpartWealthID, userUpdate.HiveID)
				updateHivequery = fmt.Sprintf("%s %s", updateHivequery, query)
				// exitingUserAnswer := user.R.ImpartWealthUserAnswers
				// for _, answer := range exitingUserAnswer {
				// 	if userHiveDemoexist[existinghiveid][uint64(answer.AnswerID)] > 0 {
				// 		userHiveDemoexist[existinghiveid][uint64(answer.AnswerID)] = userHiveDemoexist[existinghiveid][uint64(answer.AnswerID)] - 1
				// 	}
				// 	userHiveDemoexist[userUpdate.HiveID][uint64(answer.AnswerID)] = userHiveDemoexist[userUpdate.HiveID][uint64(answer.AnswerID)] + 1
				// }
				userUpdate.Users[userUpdateposition].Value = 1

				isMailSent := false
				if existinghiveid == impart.DefaultHiveID {
					isMailSent = true
				}
				if isMailSent {
					go impart.SendAWSEMails(ctx, m.db, user, impart.Hive_mail)
				}

				isNotificationEnabled := false
				if newHive != nil && newHive.NotificationTopicArn.String != "" {
					if user.R.ImpartWealthUserConfigurations != nil && !user.Admin {
						if user.R.ImpartWealthUserConfigurations[0].NotificationStatus {
							isNotificationEnabled = true
						}
					}
				}

				if existingHive != nil {
					if existingHive.NotificationTopicArn.String != "" {
						go func() {
							err := m.notificationService.UnsubscribeTopicForAllDevice(ctx, user.ImpartWealthID, existingHive.NotificationTopicArn.String)
							if err != nil {
								m.logger.Error("SubscribeTopic", zap.String("DeviceToken", existingHive.NotificationTopicArn.String),
									zap.Error(err))
							}
						}()
					}
				}
				if isNotificationEnabled {
					go func() {
						deviceDetails, devErr := m.GetUserDevices(ctx, "", user.ImpartWealthID, "")
						if devErr != nil {
							m.logger.Error("unable to find device", zap.Error(devErr))
						}
						if len(deviceDetails) > 0 {
							for _, device := range deviceDetails {
								if (device.LastloginAt == null.Time{}) {
									endpointARN, err := m.notificationService.GetEndPointArn(ctx, device.DeviceToken, "")
									if err != nil {
										m.logger.Error("End point ARN finding failed", zap.String("DeviceToken", device.DeviceToken),
											zap.Error(err))
									}
									if endpointARN != "" && newHive.NotificationTopicArn.String != "" {
										m.notificationService.SubscribeTopic(ctx, user.ImpartWealthID, newHive.NotificationTopicArn.String, endpointARN)
									}
								}
							}
						}
						if isMailSent && isNotificationEnabled {
							notificationData := impart.NotificationData{
								EventDatetime:  impart.CurrentUTC(),
								HiveID:         newHive.HiveID,
								ImpartWealthID: user.ImpartWealthID,
								Email:          user.Email,
							}
							alert := impart.Alert{
								Title: aws.String(impart.AssignHiveTitle),
								Body:  aws.String(impart.AssignHiveBody),
							}
							err := m.notificationService.Notify(ctx, notificationData, alert, user.ImpartWealthID)
							if err != nil {
								m.logger.Error("push-notification : error attempting to send hive notification ",
									zap.Any("postData", notificationData),
									zap.Any("postData", alert),
									zap.Error(err))
							}
						}
					}()
				}
			}
		} else if userUpdate.Type == impart.RemoveAdmin {
			if !user.Admin {
				userUpdate.Users[userUpdateposition].Message = "User is not an admin."
			} else {
				impartWealthIds = fmt.Sprintf("%s '%s' ,", impartWealthIds, user.ImpartWealthID)
				userUpdate.Users[userUpdateposition].Value = 1

				var existingHive *dbmodels.Hive

				for _, h := range user.R.MemberHiveHives {
					existingHive = h
				}
				deviceDetails := user.R.ImpartWealthUserDevices
				isnotificationEnabled := false
				if existingHive != nil && existingHive.NotificationTopicArn.String != "" {
					if user.R.ImpartWealthUserConfigurations != nil && !user.Admin {
						if user.R.ImpartWealthUserConfigurations[0].NotificationStatus {
							isnotificationEnabled = true
						}
					}
				}
				if isnotificationEnabled {
					// var waitGrp sync.WaitGroup
					// waitGrp.Add(1)
					go func() {
						// defer waitGrp.Done()
						for _, device := range deviceDetails {
							if (device.LastloginAt == null.Time{}) {
								endpointARN, err := m.notificationService.GetEndPointArn(ctx, device.DeviceToken, "")
								if err != nil {
									m.logger.Error("End point ARN finding failed", zap.String("DeviceToken", device.DeviceToken),
										zap.Error(err))
								}
								if endpointARN != "" {
									m.notificationService.SubscribeTopic(ctx, user.ImpartWealthID, existingHive.NotificationTopicArn.String, endpointARN)
								}
							}
						}
					}()
					// waitGrp.Wait()
				}
			}
		}
	}
	// for hive, demo := range userHiveDemoexist {updateHiveDemographic
	// 	for answer, cnt := range demo {
	// 		query := fmt.Sprintf("update hive_user_demographic set user_count=%d where hive_id=%d and answer_id=%d;", cnt, hive, answer)
	// 		updateHiveDemographic = fmt.Sprintf("%s %s", updateHiveDemographic, query)
	// 	}
	// }
	if userUpdate.Type == impart.RemoveAdmin {
		impartWealthIds = strings.Trim(impartWealthIds, ",")
		updateQuery = fmt.Sprintf(`UPDATE  user 
		SET avatar_background=(CASE CEIL(RAND()*3) 
		WHEN 1 THEN '#30A096'
		WHEN 2 THEN '#DE750D' 
		WHEN 3 THEN '#F4D304' END),
		admin=false where impart_wealth_id in(%s);`,
			impartWealthIds)
	}
	query := fmt.Sprintf("%s %s ", updateQuery, updateHivequery)
	m.logger.Info("update query", zap.String("query", query))
	_, err := queries.Raw(query).ExecContext(ctx, m.db)
	if err != nil {
		m.logger.Error("unable to excute query", zap.String("query", query),
			zap.Error(err))
		return userUpdate, err
	}
	if userUpdate.Type == impart.AddToHive || userUpdate.Type == impart.AddToWaitlist {
		_, err = userDetails.UpdateAll(ctx, m.db, dbmodels.M{"hive_updated_at": impart.CurrentUTC()})
		m.logger.Error("hive Update Failed", zap.Any("query", userDetails),
			zap.Error(err))
	}
	return userUpdate, nil
}

func (m *mysqlStore) CreateMailChimpForExistingUsers(ctx context.Context) error {
	newcluse := Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.Blocked), false)
	usersWhere := []QueryMod{
		newcluse,
		Load(dbmodels.UserRels.ImpartWealthProfile),
		Load(dbmodels.UserRels.MemberHiveHives),
		Load(dbmodels.UserRels.ImpartWealthUserAnswers),
		Load(Rels(dbmodels.UserRels.ImpartWealthUserAnswers, dbmodels.UserAnswerRels.Answer, dbmodels.AnswerRels.Question)),
	}

	out, err := dbmodels.Users(usersWhere...).All(ctx, m.db)
	status := ""
	if err != nil {
		return err
	}
	params := &members.GetParams{
		Status: members.StatusSubscribed,
	}
	cfg, _ := config.GetImpart()
	listMembers, err := members.Get(cfg.MailchimpAudienceId, params)
	if err != nil {
		return err
	}
	for _, user := range out {
		userAnswer := impart.GetUserAnswerList()
		userExist := Contains(listMembers, user.Email)
		if !userExist {
			if len(user.R.MemberHiveHives) > 0 {
				if user.R.MemberHiveHives[0].HiveID == impart.DefaultHiveID {
					status = impart.WaitList
				} else {
					status = impart.Hive
				}
			}
			if len(user.R.ImpartWealthUserAnswers) > 0 {
				for _, anser := range user.R.ImpartWealthUserAnswers {
					userAnswer[anser.R.Answer.QuestionID] = fmt.Sprintf("%s,%s", userAnswer[anser.R.Answer.QuestionID], anser.R.Answer.Text)
					userAnswer[anser.R.Answer.R.Question.QuestionID] = strings.Trim(userAnswer[anser.R.Answer.R.Question.QuestionID], ",")
				}
			}
			mergeFlds := impart.SetMailChimpAnswer(userAnswer, status, "")
			params := &members.NewParams{
				EmailAddress: user.Email,
				Status:       members.StatusSubscribed,
				MergeFields:  mergeFlds,
			}
			_, err := members.New(cfg.MailchimpAudienceId, params)
			if err != nil {
				m.logger.Info("new user requset failed in MailChimp", zap.String("updateuser", user.Email),
					zap.String("contextUser", user.ImpartWealthID))
			}
		}

	}

	return nil
}

func Contains(users *members.ListMembers, userEmail string) bool {
	for _, mail := range users.Members {
		if mail.EmailAddress == userEmail {
			return true
		}
	}
	return false
}

func (m *mysqlStore) GetUserAnswer(ctx context.Context, impartWealthId string) (dbmodels.UserAnswerSlice, error) {
	qm := []QueryMod{
		dbmodels.UserAnswerWhere.ImpartWealthID.EQ(impartWealthId),
	}
	qm = append(qm, Load(Rels(dbmodels.UserAnswerRels.Answer, dbmodels.AnswerRels.Question, dbmodels.QuestionRels.Questionnaire)))
	qm = append(qm, Load(Rels(dbmodels.UserAnswerRels.Answer, dbmodels.AnswerRels.Question, dbmodels.QuestionRels.Type)))
	qm = append(qm, Load(Rels(dbmodels.UserAnswerRels.ImpartWealth)))

	userAnswers, err := dbmodels.UserAnswers(qm...).All(ctx, m.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbmodels.UserAnswerSlice{}, impart.ErrNotFound
		}
	}
	return userAnswers, nil
}

//  GetUserDevice : Get the user device
func (m *mysqlStore) GetUserDevices(ctx context.Context, token string, impartID string, deviceToken string) (dbmodels.UserDeviceSlice, error) {
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

	device, err := dbmodels.UserDevices(where...).All(ctx, m.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return device, err
}

func (m *mysqlStore) GetHiveNotification(ctx context.Context) error {
	HiveNotificaton := impart.GetHiveNotificationDetails()
	for _, hive := range HiveNotificaton {
		startdate := impart.CurrentUTC().AddDate(0, 0, -(hive.Day + 1))
		enddate := impart.CurrentUTC().AddDate(0, 0, -(hive.Day + 2))
		dates := []time.Time{startdate, enddate}
		userList, err := m.getUserAll(ctx, nil, false, 2, impart.DefaultHiveID, dates)
		if err != nil {
			m.logger.Error("User fetching for notification error-", zap.Any("error", err))
			return err
		}
		notificationData := impart.NotificationData{
			EventDatetime: impart.CurrentUTC(),
			Path:          hive.Redirection,
		}
		for _, user := range userList {
			isNotificationEnabled := false
			if user.R.ImpartWealthUserConfigurations != nil && !user.Admin {
				if user.R.ImpartWealthUserConfigurations[0].NotificationStatus {
					isNotificationEnabled = true
				}
			}
			if !isNotificationEnabled {
				break
			}
			body := hive.Body
			if hive.IncludeFirstName {
				body = fmt.Sprintf(hive.Body, user.FirstName)
			}
			alert := impart.Alert{
				Title: aws.String(hive.Title),
				Body:  aws.String(body),
			}
			m.logger.Info("User Details",
				zap.Any("User", user),
				zap.Any("notificationData", notificationData),
				zap.Any("alert", alert))

			m.notificationService.Notify(context.TODO(), notificationData, alert, user.ImpartWealthID)
		}
	}
	return nil
}
