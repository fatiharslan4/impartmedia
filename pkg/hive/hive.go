package hive

import (
	"context"
	"fmt"
	"strings"

	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
)

// VoteInput is the input to register an upvote or downvote on a comment or post.
type VoteInput struct {
	PostID, CommentID uint64
	Upvote            bool
	Increment         bool
}

func (s *service) Votes(ctx context.Context, v VoteInput) (models.PostCommentTrack, impart.Error) {
	var out models.PostCommentTrack
	s.logger.Debug("received vote request", zap.Any("input", v))

	var in data.ContentInput

	if v.CommentID > 0 {
		in.Id = v.CommentID
		in.Type = data.Comment
	} else {
		in.Id = v.PostID
		in.Type = data.Post
	}
	var err error
	var actionType types.Type
	if v.Upvote {
		if v.Increment {
			err = s.reactionData.AddUpVote(ctx, in)
			actionType = types.UpVote
		} else {
			err = s.reactionData.TakeUpVote(ctx, in)
			actionType = types.TakeUpVote
		}
	} else { //Downvote
		if v.Increment {
			err = s.reactionData.AddDownVote(ctx, in)
			actionType = types.DownVote
		} else {
			err = s.reactionData.TakeDownVote(ctx, in)
			actionType = types.TakeDownVote
		}
	}

	if err != nil {
		s.logger.Error("error on vote", zap.Error(err), zap.Any("vote", v))
	} else {
		// send notification on up,down,take votes
		err = s.SendNotificationOnVote(ctx, actionType, v, in)
		if err != nil {
			s.logger.Error("error on vote notification", zap.Error(err), zap.Any("vote", v))
		}
	}
	out, err = s.reactionData.GetUserTrack(ctx, in)
	if err != nil {
		s.logger.Error("error getting updated tracked item track store", zap.Error(err), zap.Any("vote", v))
		return out, impart.NewError(err, "unable to retrieve recently tracked content")
	}

	return out, nil
}

func (s *service) Logger() *zap.Logger {
	return s.logger
}

func (s *service) sendNotification(data impart.NotificationData, alert impart.Alert, impartWealthId string) error {
	return s.notificationService.Notify(context.TODO(), data, alert, impartWealthId)
}

// REturns unauthorized if
func (s *service) validateHiveAccess(ctx context.Context, hiveID uint64) impart.Error {
	ctxUser := impart.GetCtxUser(ctx)
	if ctxUser == nil {
		return impart.NewError(impart.ErrUnauthorized, "user is not a member of this hive")
	}
	if impart.GetCtxUser(ctx).Admin {
		return nil
	}

	for _, h := range ctxUser.R.MemberHiveHives {
		if h.HiveID == hiveID {
			return nil
		}
	}
	return impart.NewError(impart.ErrUnauthorized, "user is not a member of this hive")

}

func (s *service) GetHive(ctx context.Context, hiveID uint64) (models.Hive, impart.Error) {
	if err := s.validateHiveAccess(ctx, hiveID); err != nil {
		return models.Hive{}, err
	}

	dbHive, err := s.hiveData.GetHive(ctx, hiveID)
	if err != nil {
		s.logger.Error("error getting hive", zap.Error(err))
		if err == impart.ErrNotFound {
			return models.Hive{}, impart.NewError(err, fmt.Sprintf("hive %v not found", hiveID))
		}
		return models.Hive{}, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to retrieve hive %v", hiveID))
	}

	hive, err := models.HiveFromDB(dbHive)
	if err != nil {
		s.logger.Error("couldn't convert db model to hive", zap.Error(err))
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "bad db model")
	}
	return hive, nil
}

// If the auth user is an admin, then return all hives.  Otherwise only return hives the user is a member of.
func (s *service) GetHives(ctx context.Context) (models.Hives, impart.Error) {
	var err error

	dbHives, err := s.hiveData.GetHives(ctx)
	if err != nil {
		return models.Hives{}, impart.NewError(impart.ErrUnknown, "unable fetch dbmodels")
	}

	hives, err := models.HivesFromDB(dbHives)
	if err != nil {
		return models.Hives{}, impart.NewError(impart.ErrUnknown, "unable to convert hives from dbmodel")
	}

	return hives, nil
}

func (s *service) CreateHive(ctx context.Context, hive models.Hive) (models.Hive, impart.Error) {
	var err error

	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.SuperAdmin {
		return models.Hive{}, impart.NewError(impart.ErrUnauthorized, "non-admin users cannot create hives.")
	}
	if len(strings.TrimSpace(hive.HiveName)) < 3 {
		return models.Hive{}, impart.NewError(impart.ErrBadRequest, "Hivename must be greater than or equal to 3.")
	}
	if len(strings.TrimSpace(hive.HiveName)) > 60 {
		return models.Hive{}, impart.NewError(impart.ErrBadRequest, "Hivename must be less than or equal to 60.")
	}

	dbh, err := hive.ToDBModel()
	if err != nil {
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "unable to convert hives to  dbmodel")
	}
	dbh, err = s.hiveData.NewHive(ctx, dbh)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return hive, impart.NewError(impart.ErrUnknown, "Hive name already exists.", impart.HiveID)
		}
		return hive, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to create hive %s", hive.HiveName), impart.HiveID)
	}

	cfg, _ := config.GetImpart()
	topicInput := fmt.Sprintf("SNSHiveNotification-%s-%d", cfg.Env, dbh.HiveID)
	topicInput = strings.Replace(topicInput, " ", "-", -1)
	s.logger.Info("Topic", zap.Any("topicInput", topicInput))
	topic, err := s.notificationService.CreateNotificationTopic(ctx, topicInput)
	if err != nil {
		s.logger.Error("error creating hive topic", zap.Error(err))
	}
	s.logger.Info("Topic details", zap.Any("topic", topic))
	if topic != nil {
		s.logger.Info("Topic details is not null", zap.Any("topicARn", topic.TopicArn))
		dbh.NotificationTopicArn = null.StringFrom(*topic.TopicArn)
		if _, err = dbh.Update(ctx, s.db, boil.Infer()); err != nil {
			s.logger.Error("Topic details update failed in Db", zap.Error(err))
		}
	}

	out, err := models.HiveFromDB(dbh)
	if err != nil {
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "unable to convert hives to  dbmodel")
	}

	return out, nil
}

func (s *service) EditHive(ctx context.Context, hive models.Hive) (models.Hive, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.SuperAdmin {
		return models.Hive{}, impart.NewError(impart.ErrUnauthorized, "non-admin users cannot create hives.")
	}
	if len(strings.TrimSpace(hive.HiveName)) < 3 {
		return models.Hive{}, impart.NewError(impart.ErrBadRequest, "Hivename must be greater than or equal to 3.")
	}
	if len(strings.TrimSpace(hive.HiveName)) > 60 {
		return models.Hive{}, impart.NewError(impart.ErrBadRequest, "Hivename must be less than or equal to 60.")
	}
	dbh, err := s.hiveData.EditHive(ctx, hive)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return hive, impart.NewError(impart.ErrUnknown, "Hive name already exists.", impart.HiveID)
		}
		return hive, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to create hive %s", hive.HiveName))
	}
	out, err := models.HiveFromDB(dbh)
	if err != nil {
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "unable to convert hives to  dbmodel")
	}

	return out, nil
}

// SendNotificationOnVote
// vote may be to post or comment under post
func (s *service) SendNotificationOnVote(ctx context.Context, actionType types.Type, v VoteInput, in data.ContentInput) error {
	var err error
	// check the type is comment
	if in.Type == data.Comment {
		err = s.SendCommentNotification(models.CommentNotificationInput{
			Ctx:        ctx,
			CommentID:  in.Id,
			ActionType: actionType,
			ActionData: "",
		})
	}

	// check the type is post
	if in.Type == data.Post {
		err = s.SendPostNotification(models.PostNotificationInput{
			Ctx:        ctx,
			PostID:     in.Id,
			ActionType: actionType,
			ActionData: "",
		})
	}

	return err
}

// Get Reported User
// get post reported users list
func (s *service) GetReportedUser(ctx context.Context, posts models.Posts) (models.Posts, error) {
	return s.postData.GetReportedUser(ctx, posts)
}

func (s *service) GetReportedContents(ctx context.Context, gpi data.GetReportedContentInput) (models.PostComments, *models.NextPage, error) {
	return s.hiveData.GetReportedContents(ctx, gpi)
}

func (s *service) DeleteHive(ctx context.Context, hiveID uint64) impart.Error {
	if hiveID == impart.DefaultHiveID {
		return impart.NewError(impart.ErrBadRequest, "You cannot delete the default hive.")
	}
	ctxUser := impart.GetCtxUser(ctx)
	_, err := s.hiveData.GetHive(ctx, hiveID)
	if err != nil {
		s.logger.Error("error fetching hive trying to edit", zap.Error(err))
		return impart.NewError(impart.ErrBadRequest, "Unable to find the hive.")
	}
	clientId := impart.GetCtxClientID(ctx)
	if clientId == impart.ClientId {
		if !ctxUser.SuperAdmin {
			return impart.NewError(impart.ErrUnauthorized, "Only super admin can delete hive.")
		}
	} else {
		return impart.NewError(impart.ErrUnauthorized, "You have no permisson to delete hive")
	}
	err = s.hiveData.DeleteHive(ctx, hiveID)
	if err != nil {
		return impart.UnknownError
	}

	return nil
}

func (s *service) HiveBulkOperations(ctx context.Context, hiveUpdates models.HiveUpdate) *models.HiveUpdate {
	hiveOutput := models.HiveUpdate{}
	hiveDatas := make([]models.HiveData, len(hiveUpdates.Hives), len(hiveUpdates.Hives))
	hiveOutput.Action = hiveUpdates.Action
	HiveIds := make([]interface{}, 0, len(hiveUpdates.Hives))
	for i, hive := range hiveUpdates.Hives {
		hives := &models.HiveData{}
		hives.HiveID = hive.HiveID
		hives.Name = hive.Name
		hives.Message = "No delete activity."
		hives.Status = false
		if hive.HiveID > 0 {
			HiveIds = append(HiveIds, (hive.HiveID))
		}
		hiveDatas[i] = *hives
	}
	hiveOutput.Hives = hiveDatas
	hiveOutputRslt := &hiveOutput

	hivesop, err := s.hiveData.GetHiveFromList(ctx, HiveIds)

	if err != nil || len(hivesop) == 0 {
		return hiveOutputRslt
	}
	err = s.hiveData.DeleteBulkHive(ctx, hivesop)
	if err != nil {
		return hiveOutputRslt
	}
	lenhive := len(hiveOutputRslt.Hives)
	for _, hive := range hivesop {
		for cnt := 0; cnt < lenhive; cnt++ {
			if hiveOutputRslt.Hives[cnt].HiveID == hive.HiveID && hive.HiveID != impart.DefaultHiveID {
				hiveOutputRslt.Hives[cnt].Message = "Hive deleted."
				hiveOutputRslt.Hives[cnt].Status = true
				break
			}
		}
	}
	return hiveOutputRslt
}
