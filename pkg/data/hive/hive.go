package data

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/beeker1121/mailchimp-go/lists/members"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/media"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

var _ Hives = &mysqlHiveData{}
var _ HiveService = &mysqlHiveData{}

type mysqlHiveData struct {
	logger       *zap.Logger
	db           *sql.DB
	MediaStorage media.StorageConfigurations
}

//counterfeiter:generate . HiveService
type HiveService interface {
	Hives
	Comments
	Posts
	UserTrack
}

func NewHiveService(db *sql.DB, logger *zap.Logger) HiveService {
	return &mysqlHiveData{
		logger: logger,
		db:     db,
	}
}

// Hives is the interface for Hive CRUD operations
type Hives interface {
	GetHives(ctx context.Context) (dbmodels.HiveSlice, error)
	GetHive(ctx context.Context, hiveID uint64) (*dbmodels.Hive, error)
	NewHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error)
	EditHive(ctx context.Context, hive models.Hive) (*dbmodels.Hive, error)
	PinPost(ctx context.Context, hiveID, postID uint64, pin bool, isAdminActivity bool) error
	GetReviewedPosts(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetUnreviewedReportedPosts(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetPostsWithUnreviewedComments(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetPostsWithReviewedComments(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetReportedContents(ctx context.Context, getInput GetReportedContentInput) (models.PostComments, *models.NextPage, error)
	DeleteHive(ctx context.Context, hiveID uint64) error
	GetHiveFromList(ctx context.Context, hiveIds []interface{}) (dbmodels.HiveSlice, error)
	DeleteBulkHive(ctx context.Context, hiveIDs dbmodels.HiveSlice) error
	PinPostForBulkPostAction(ctx context.Context, postHive map[uint64]uint64, pin bool, isAdminActivity bool) error
	NewHiveRule(ctx context.Context, hiverule *dbmodels.HiveRule, hiveCriteria dbmodels.HiveRulesCriteriumSlice) (*dbmodels.HiveRule, error)
	EditHiveRule(ctx context.Context, hiverule models.HiveRule) (*dbmodels.HiveRule, impart.Error)
}

func (d *mysqlHiveData) GetHives(ctx context.Context) (dbmodels.HiveSlice, error) {
	ctxUser := impart.GetCtxUser(ctx)
	if ctxUser == nil {
		return dbmodels.HiveSlice{}, impart.UnknownError
	}
	if !ctxUser.Admin {
		return ctxUser.R.MemberHiveHives, nil
	}
	return dbmodels.Hives().All(ctx, d.db)
}

func (d *mysqlHiveData) GetHive(ctx context.Context, hiveID uint64) (*dbmodels.Hive, error) {
	return dbmodels.FindHive(ctx, d.db, hiveID)
}

func (d *mysqlHiveData) NewHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error) {
	// ctxUser := impart.GetCtxUser(ctx)
	// if !ctxUser.SuperAdmin {
	// 	return nil, impart.ErrUnauthorized
	// }
	if err := hive.Insert(ctx, d.db, boil.Infer()); err != nil {
		d.logger.Error("Hive creation failed", zap.Error(err))
		return nil, err
	}
	queryFirst := `
					INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
					SELECT ?,question_id,answer_id,0
					FROM answer;`
	_, err := queries.Raw(queryFirst, hive.HiveID).ExecContext(ctx, d.db)
	if err != nil {
		fmt.Println(err)
		d.logger.Error("Updating Hive demographic data failed", zap.String("Hive name", hive.Name))
	}
	return hive, hive.Reload(ctx, d.db)
}

func (d *mysqlHiveData) EditHive(ctx context.Context, hive models.Hive) (*dbmodels.Hive, error) {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.SuperAdmin {
		return nil, impart.ErrUnauthorized
	}
	existing, err := dbmodels.FindHive(ctx, d.db, hive.HiveID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.ErrNotFound
		}
		return nil, err
	}

	if existing.Name != hive.HiveName && hive.HiveName != "" {
		existing.Name = hive.HiveName
	}

	if _, err := existing.Update(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}

	return existing, existing.Reload(ctx, d.db)
}

// PinPost takes a hive and post id of a post ot pin or unpin
// if a post is being pinned, within the same transaction we need to (maybe) unpin the old post,
// mark the new post as pinned, and update the hive to point to the new post.
func (d *mysqlHiveData) PinPost(ctx context.Context, hiveID, postID uint64, pin bool, isAdminActivity bool) error {
	// ctxUser := impart.GetCtxUser(ctx)
	if !isAdminActivity {
		return impart.ErrUnauthorized
	}
	toPin, err := dbmodels.FindPost(ctx, d.db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
		return err
	}
	hive, err := dbmodels.Hives(dbmodels.HiveWhere.HiveID.EQ(hiveID)).One(ctx, d.db)
	if err != nil {
		return err
	}
	if toPin.Pinned == pin && hive.PinnedPostID.Valid && hive.PinnedPostID.Uint64 == postID {
		return impart.NewError(impart.ErrBadRequest, "Resource already matches exactly as request.")
	}
	if toPin.HiveID != hiveID {
		return impart.NewError(impart.ErrBadRequest, "HiveId not matching.")
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer impart.CommitRollbackLogger(tx, err, d.logger)

	if hive.PinnedPostID.Valid {
		if !pin && postID == hive.PinnedPostID.Uint64 || pin {
			if hive.PinnedPostID.Uint64 > 0 {
				existingPinnedPost, err := dbmodels.FindPost(ctx, d.db, hive.PinnedPostID.Uint64)
				if err != nil {
					return err
				}
				existingPinnedPost.Pinned = false
				if _, err := existingPinnedPost.Update(ctx, d.db, boil.Whitelist(dbmodels.PostColumns.Pinned)); err != nil {
					return err
				}
			}
		}
	}

	toPin.Pinned = pin
	_, err = toPin.Update(ctx, tx, boil.Whitelist(dbmodels.PostColumns.Pinned))

	if pin {
		hive.PinnedPostID.SetValid(postID)
		_, err = hive.Update(ctx, tx, boil.Whitelist(dbmodels.HiveColumns.PinnedPostID))
		return err
	} else {
		//unpin
		hive.PinnedPostID = null.Uint64{}
		_, err = hive.Update(ctx, tx, boil.Whitelist(dbmodels.HiveColumns.PinnedPostID))
		return err
	}
}

func (d *mysqlHiveData) DeleteHive(ctx context.Context, hiveID uint64) error {
	hives, err := dbmodels.FindHive(ctx, d.db, hiveID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	var memberHives []models.MemberHive
	err = queries.Raw(`
	SELECT member_hive_id,user.impart_wealth_id  FROM hive_members 
	join user on hive_members.member_impart_wealth_id=user.impart_wealth_id
	where user.deleted_at is null and member_hive_id=?
	`, hiveID).Bind(ctx, d.db, &memberHives)

	hives.Name = fmt.Sprintf("%s-%d-%s", hives.Name, hives.HiveID, "Deleted")
	if _, err := hives.Update(ctx, d.db, boil.Infer()); err != nil {
		return err
	}
	if _, err = hives.Delete(ctx, d.db, false); err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
		return err
	}
	updatememberhives := ""
	updateHiveDemographic := ""
	userHiveDemo := make(map[uint64]map[uint64]int)
	for _, member := range memberHives {
		query := fmt.Sprintf("update hive_members set member_hive_id=%d where member_impart_wealth_id='%s';", impart.DefaultHiveID, member.ImpartWealthID)
		updatememberhives = fmt.Sprintf("%s %s", updatememberhives, query)
	}
	hive_lst := []uint64{hiveID, impart.DefaultHiveID}
	answer, err := dbmodels.HiveUserDemographics(
		dbmodels.HiveUserDemographicWhere.HiveID.IN(hive_lst),
	).All(ctx, d.db)

	for _, demohive := range answer {
		data := userHiveDemo[uint64(demohive.HiveID)]
		if data == nil {
			count := make(map[uint64]int)
			count[uint64(demohive.AnswerID)] = int(demohive.UserCount)
			userHiveDemo[uint64(demohive.HiveID)] = count
		} else {
			data[uint64(demohive.AnswerID)] = int(demohive.UserCount)
		}
	}
	for _, demohive := range answer {
		if demohive.HiveID == hiveID {
			userHiveDemo[impart.DefaultHiveID][uint64(demohive.AnswerID)] = userHiveDemo[impart.DefaultHiveID][uint64(demohive.AnswerID)] + userHiveDemo[demohive.HiveID][uint64(demohive.AnswerID)]
			userHiveDemo[demohive.HiveID][uint64(demohive.AnswerID)] = 0
		}
	}
	for hive, demo := range userHiveDemo {
		for answer, cnt := range demo {
			query := fmt.Sprintf("update hive_user_demographic set user_count=%d where hive_id=%d and answer_id=%d;", cnt, hive, answer)
			updateHiveDemographic = fmt.Sprintf("%s %s", updateHiveDemographic, query)
		}
	}

	if updateHiveDemographic != "" || updatememberhives != "" {
		query := fmt.Sprintf("%s %s", updateHiveDemographic, updatememberhives)
		_, _ = queries.Raw(query).ExecContext(ctx, d.db)
	}

	return nil
}

func (d *mysqlHiveData) DeleteBulkHive(ctx context.Context, hiveInput dbmodels.HiveSlice) error {
	updateQuery := ""
	updatememberHive := ""
	var allUser []string
	currTime := time.Now().In(boil.GetLocation())
	golangDateTime := currTime.Format("2006-01-02 15:04:05.000")

	for _, hive := range hiveInput {
		if hive.HiveID == impart.DefaultHiveID {
			continue
		}
		deleteName := fmt.Sprintf("%s-%d-%s", hive.Name, hive.HiveID, "Deleted")
		query := fmt.Sprintf("Update hive set deleted_at='%s' , name='%s' where hive_id='%d';", golangDateTime, deleteName, hive.HiveID)
		updateQuery = fmt.Sprintf("%s %s", updateQuery, query)
		exitingmembers := hive.R.MemberImpartWealthUsers
		for _, member := range exitingmembers {
			query := fmt.Sprintf("update hive_members set member_hive_id=%d where member_impart_wealth_id='%s';", impart.DefaultHiveID, member.ImpartWealthID)
			updatememberHive = fmt.Sprintf("%s %s", updatememberHive, query)
			allUser = append(allUser, member.Email)
		}
	}
	query := fmt.Sprintf("%s %s ", updateQuery, updatememberHive)
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer impart.CommitRollbackLogger(tx, err, d.logger)
	_, err = queries.Raw(query).ExecContext(ctx, d.db)
	if err != nil {
		return err
	}
	// Update mailChimp
	go func() {
		cfg, _ := config.GetImpart()
		for hiveUser := range allUser {
			mailChimpParams := &members.UpdateParams{
				MergeFields: map[string]interface{}{"STATUS": impart.WaitList},
			}
			_, err = members.Update(cfg.MailchimpAudienceId, allUser[hiveUser], mailChimpParams)
			if err != nil {
				d.logger.Error("Delete user requset failed in MailChimp", zap.String("deleteUser", allUser[hiveUser]),
					zap.String("contextUser", allUser[hiveUser]))
			}
		}
	}()
	go impart.UserDemographicsUpdate(ctx, d.db, true, true)
	return nil
}

func rollbackIfError(tx *sql.Tx, err error, logger *zap.Logger) error {
	rErr := tx.Rollback()
	if rErr != nil {
		logger.Error("unable to rollback transaction", zap.Error(rErr))
		return fmt.Errorf(rErr.Error(), err)
	}
	return err
}

func (d *mysqlHiveData) GetHiveFromList(ctx context.Context, hiveIds []interface{}) (dbmodels.HiveSlice, error) {
	orderByMod := qm.WhereIn("hive_id in ?", hiveIds...)
	queryMods := []qm.QueryMod{
		orderByMod,
		qm.Load(dbmodels.HiveRels.MemberImpartWealthUsers),
		qm.Load(dbmodels.HiveRels.HiveUserDemographics),
		qm.Load(dbmodels.HiveRels.AdminImpartWealthUsers),
	}
	hives, err := dbmodels.Hives(queryMods...).All(ctx, d.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return hives, err
}

// PinPost takes a hive and post id of a post ot pin or unpin
// if a post is being pinned, within the same transaction we need to (maybe) unpin the old post,
// mark the new post as pinned, and update the hive to point to the new post.
func (d *mysqlHiveData) PinPostForBulkPostAction(ctx context.Context, postHiveDetails map[uint64]uint64, pin bool, isAdminActivity bool) error {
	// ctxUser := impart.GetCtxUser(ctx)
	if !isAdminActivity {
		return impart.ErrUnauthorized
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer impart.CommitRollbackLogger(tx, err, d.logger)

	query := "UPDATE post JOIN hive ON post.post_id = hive.pinned_post_id SET post.pinned=false WHERE hive.hive_id in("
	for i := range postHiveDetails {
		qry := fmt.Sprintf("%d,", i)
		query = fmt.Sprintf("%s %s", query, qry)
	}
	query = strings.Trim(query, ",")
	query = fmt.Sprintf("%s );", query)

	postQuery := "UPDATE hive JOIN post ON post.hive_id = hive.hive_id SET hive.pinned_post_id=post.post_id WHERE post.post_id in("
	for _, post := range postHiveDetails {
		qry := fmt.Sprintf("%d,", post)
		postQuery = fmt.Sprintf("%s %s", postQuery, qry)
	}
	postQuery = strings.Trim(postQuery, ",")
	postQuery = fmt.Sprintf("%s );", postQuery)

	finlQuery := fmt.Sprintf("%s %s", query, postQuery)

	_, err = queries.Raw(finlQuery).QueryContext(ctx, d.db)
	if err != nil {
		d.logger.Error("error attempting to creating bulk post  data tag ", zap.Any("post", finlQuery), zap.Error(err))
		return err
	}

	return nil

}

func (d *mysqlHiveData) NewHiveRule(ctx context.Context, hiveRule *dbmodels.HiveRule, hiveCriteria dbmodels.HiveRulesCriteriumSlice) (*dbmodels.HiveRule, error) {
	if err := hiveRule.Insert(ctx, d.db, boil.Infer()); err != nil {
		d.logger.Error("HiveRule creation failed", zap.Error(err))
		return nil, err
	}
	err := hiveRule.AddRuleHiveRulesCriteria(ctx, d.db, true, hiveCriteria...)
	if err != nil {
		d.logger.Error("HiveRule criteria creation failed", zap.Error(err))
	}
	if (hiveRule.HiveID != null.Uint64{}) && hiveRule.HiveID.Uint64 > 0 {
		newHive := &dbmodels.Hive{HiveID: hiveRule.HiveID.Uint64}
		errHive := hiveRule.AddHives(ctx, d.db, false, newHive)
		if errHive != nil {
			d.logger.Error("New hive rule map failed", zap.Any("hive", newHive),
				zap.Error(errHive))
		}
	}
	return hiveRule, hiveRule.Reload(ctx, d.db)
}

func (d *mysqlHiveData) EditHiveRule(ctx context.Context, hiveRule models.HiveRule) (*dbmodels.HiveRule, impart.Error) {

	existing, err := dbmodels.FindHiveRule(ctx, d.db, hiveRule.RuleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, impart.NewError(impart.ErrNotFound, string(impart.HiveRuleNotExist))
		}
		return nil, impart.NewError(impart.ErrBadRequest, string(impart.HiveRuleFetchingFailed))
	}
	if existing.Status == hiveRule.Status {
		return nil, impart.NewError(impart.ErrBadRequest, string(impart.HiveRuleSameStatus))
	}
	existing.Status = hiveRule.Status
	if _, err := existing.Update(ctx, d.db, boil.Infer()); err != nil {
		return nil, impart.NewError(impart.ErrBadRequest, string(impart.HiveRuleUpdateFailed))
	}

	return existing, nil
}
