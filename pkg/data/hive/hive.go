package data

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/media"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	EditHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error)
	PinPost(ctx context.Context, hiveID, postID uint64, pin bool) error
	GetReviewedPosts(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetUnreviewedReportedPosts(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetPostsWithUnreviewedComments(ctx context.Context, hiveId uint64, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetPostsWithReviewedComments(ctx context.Context, hiveId uint64, reviewDate time.Time, offset int) (dbmodels.PostSlice, models.NextPage, error)
	GetReportedContents(ctx context.Context, getInput GetReportedContentInput) (models.PostComments, *models.NextPage, error)
	DeleteHive(ctx context.Context, hiveID uint64) error
	GetHiveFromList(ctx context.Context, hiveIds []interface{}) (dbmodels.HiveSlice, error)
	DeleteBulkHive(ctx context.Context, hiveIDs dbmodels.HiveSlice) error
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
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
		return nil, impart.ErrUnauthorized
	}

	// hive.HiveID = 0
	if err := hive.Insert(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}

	return hive, hive.Reload(ctx, d.db)
}

func (d *mysqlHiveData) EditHive(ctx context.Context, hive *dbmodels.Hive) (*dbmodels.Hive, error) {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
		return nil, impart.ErrUnauthorized
	}

	if _, err := hive.Update(ctx, d.db, boil.Infer()); err != nil {
		return nil, err
	}

	return hive, hive.Reload(ctx, d.db)
}

// PinPost takes a hive and post id of a post ot pin or unpin
// if a post is being pinned, within the same transaction we need to (maybe) unpin the old post,
// mark the new post as pinned, and update the hive to point to the new post.
func (d *mysqlHiveData) PinPost(ctx context.Context, hiveID, postID uint64, pin bool) error {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
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
	p, err := dbmodels.FindHive(ctx, d.db, hiveID)
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

	if _, err = p.Delete(ctx, d.db, false); err != nil {
		if err == sql.ErrNoRows {
			return impart.ErrNotFound
		}
		return err
	}
	for _, i := range memberHives {
		q := `
		update  hive_members set member_hive_id= ? where 
		member_impart_wealth_id= ?`
		_, err = queries.Raw(q, impart.DefaultHiveID, i.ImpartWealthID).ExecContext(ctx, d.db)
		if err != nil {
		}
	}

	answer, err := dbmodels.HiveUserDemographics(
		dbmodels.HiveUserDemographicWhere.HiveID.EQ(hiveID),
	).All(ctx, d.db)
	answerIds := make([]uint, len(answer))
	for i, a := range answer {
		if a.UserCount > 0 {
			answerIds[i] = a.AnswerID
		}
	}
	err = d.UpdateHiveUserDemographic(ctx, answerIds, true, impart.DefaultHiveID)
	err = d.UpdateHiveUserDemographic(ctx, answerIds, false, hiveID)
	return nil
}

func (m *mysqlHiveData) UpdateHiveUserDemographic(ctx context.Context, answerIds []uint, status bool, hiveId uint64) error {
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

func (d *mysqlHiveData) DeleteBulkHive(ctx context.Context, hiveInput dbmodels.HiveSlice) error {
	updateQuery := ""
	updatememberHive := ""
	updateHiveDemographic := ""
	currTime := time.Now().In(boil.GetLocation())
	golangDateTime := currTime.Format("2006-01-02 15:04:05.000")
	userHiveDemo := make(map[uint64]map[uint64]int)
	dbhiveUserDemographic, _ := dbmodels.HiveUserDemographics().All(ctx, d.db)
	for _, p := range dbhiveUserDemographic {
		data := userHiveDemo[uint64(p.HiveID)]
		if data == nil {
			count := make(map[uint64]int)
			count[uint64(p.AnswerID)] = int(p.UserCount)
			userHiveDemo[uint64(p.HiveID)] = count
		} else {
			data[uint64(p.AnswerID)] = int(p.UserCount)
		}
	}

	for _, hive := range hiveInput {
		if hive.HiveID == impart.DefaultHiveID {
			continue
		}
		query := fmt.Sprintf("Update hive set deleted_at='%s' where hive_id='%s';", golangDateTime, hive.HiveID)
		updateQuery = fmt.Sprintf("%s %s", updateQuery, query)
		exitingmembers := hive.R.MemberImpartWealthUsers
		for _, member := range exitingmembers {
			query := fmt.Sprintf("update hive_members set member_hive_id=%d where member_impart_wealth_id='%s';", impart.DefaultHiveID, member.ImpartWealthID)
			updatememberHive = fmt.Sprintf("%s %s", updatememberHive, query)
		}
		for _, hiveDemo := range dbhiveUserDemographic {
			if hiveDemo.HiveID == hive.HiveID {
				userHiveDemo[impart.DefaultHiveID][uint64(hiveDemo.AnswerID)] = userHiveDemo[impart.DefaultHiveID][uint64(hiveDemo.AnswerID)] + userHiveDemo[hive.HiveID][uint64(hiveDemo.AnswerID)]
				userHiveDemo[hive.HiveID][uint64(hiveDemo.AnswerID)] = 0
			}
		}
	}
	for hive, demo := range userHiveDemo {
		for answer, cnt := range demo {
			query := fmt.Sprintf("update hive_user_demographic set user_count=%d where hive_id=%d and answer_id=%d;", cnt, hive, answer)
			updateHiveDemographic = fmt.Sprintf("%s %s", updateHiveDemographic, query)
		}
	}
	query := fmt.Sprintf("%s %s %s", updateQuery, updatememberHive, updateHiveDemographic)
	_, err := queries.Raw(query).ExecContext(ctx, d.db)
	if err != nil {
		return err
	}
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
	clause := WhereIn(fmt.Sprintf("%s in ?", dbmodels.HiveColumns.HiveID), hiveIds...)
	usersWhere := []QueryMod{
		clause,
		Load(dbmodels.HiveRels.HiveUserDemographics),
		Load(dbmodels.HiveRels.MemberImpartWealthUsers),
	}

	u, err := dbmodels.Hives(usersWhere...).All(ctx, d.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, err
}
