package profile

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
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

func (m *mysqlStore) DeleteProfile(ctx context.Context, impartWealthID string) error {
	u, err := dbmodels.FindUser(ctx, m.db, impartWealthID)
	if err == sql.ErrNoRows || u == nil {
		return impart.ErrNotFound
	}
	if err != nil {
		return err
	}
	_, err = u.Delete(ctx, m.db, false)
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

/**
 *
 * GetUserDevice : Get the user device
 *
 */
func (m *mysqlStore) GetUserDevice(ctx context.Context, token []byte, impartID string) (*dbmodels.UserDevice, error) {
	var clause QueryMod
	if impartID != "" {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserDeviceColumns.ImpartWealthID), impartID)
	} else {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserDeviceColumns.Token), token)
	}
	where := []QueryMod{
		clause,
		Load(dbmodels.UserDeviceRels.ImpartWealth),
	}

	device, err := dbmodels.UserDevices(where...).One(ctx, m.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return device, err
}

/**
 *
 * CreateUserDevice
 *
 */

func (m *mysqlStore) CreateUserDevice(ctx context.Context, device *dbmodels.UserDevice) (*dbmodels.UserDevice, error) {
	if device == nil {
		m.logger.Error("device is nil")
		return nil, impart.ErrBadRequest
	}
	uuid := uuid.New()
	device.Token = []byte(uuid.String())

	err := device.Insert(ctx, m.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	return m.GetUserDevice(ctx, device.Token, "")
}
