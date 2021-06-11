package hive

import (
	"context"
	"fmt"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
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
	}

	// send notification on up,down,take votes
	err = s.SendNotificationOnVote(ctx, actionType, v, in)
	if err != nil {
		s.logger.Error("error on vote notification", zap.Error(err), zap.Any("vote", v))
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
	if !ctxUser.Admin {
		return models.Hive{}, impart.NewError(impart.ErrUnauthorized, "non-admin users cannot create hives.")
	}

	dbh, err := hive.ToDBModel()
	if err != nil {
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "unable to convert hives to  dbmodel")
	}

	dbh, err = s.hiveData.NewHive(ctx, dbh)
	if err != nil {
		return hive, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to create hive %s", hive.HiveName), impart.HiveID)
	}
	out, err := models.HiveFromDB(dbh)
	if err != nil {
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "unable to convert hives to  dbmodel")
	}

	return out, nil
}

func (s *service) EditHive(ctx context.Context, hive models.Hive) (models.Hive, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
		return models.Hive{}, impart.NewError(impart.ErrUnauthorized, "non-admin users cannot create hives.")
	}

	dbh, err := hive.ToDBModel()
	if err != nil {
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "unable to convert hives to  dbmodel")
	}

	dbh, err = s.hiveData.EditHive(ctx, dbh)
	if err != nil {
		return hive, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to create hive %s", hive.HiveName))
	}
	out, err := models.HiveFromDB(dbh)
	if err != nil {
		return models.Hive{}, impart.NewError(impart.ErrUnknown, "unable to convert hives to  dbmodel")
	}

	return out, nil
}

/**
 * SendNotificationOnVote
 *
 * vote may be to post or comment under post
 */
func (s *service) SendNotificationOnVote(ctx context.Context, actionType types.Type, v VoteInput, in data.ContentInput) error {
	var err error
	// check the type is comment
	if in.Type == data.Comment {
		err = s.SendCommentNotification(models.CommentNotificationInput{
			Ctx:             ctx,
			CommentID:       in.Id,
			ActionType:      actionType,
			ActionData:      "",
			NotifyPostOwner: true,
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

/**
 * Get Reported User
 *
 * get post reported users list
 */
func (s *service) GetReportedUser(ctx context.Context, posts models.Posts) (models.Posts, error) {
	return s.postData.GetReportedUser(ctx, posts)
}

func (s *service) GetReportedContents(ctx context.Context, gpi data.GetPostsInput) (models.PostComments, *models.NextPage, error) {
	return s.hiveData.GetReportedContents(ctx, gpi)
}
