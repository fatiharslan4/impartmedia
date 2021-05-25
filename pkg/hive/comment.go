package hive

import (
	"context"
	"fmt"
	"strings"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

func (s *service) GetComments(ctx context.Context, postID uint64, limit, offset int) (models.Comments, *models.NextPage, impart.Error) {
	dbComments, nextPage, err := s.commentData.GetComments(ctx, postID, limit, offset)
	if err != nil {
		if err == impart.ErrNotFound {
			return models.Comments{}, nil, impart.NewError(err, "no comments found for post")
		} else {
			return models.Comments{}, nil, impart.NewError(impart.ErrUnknown, "couldn't fetch comments")
		}
	}
	return models.CommentsFromDBModelSlice(dbComments), nextPage, nil
}

func (s *service) GetComment(ctx context.Context, commentID uint64) (models.Comment, impart.Error) {
	dbComment, err := s.commentData.GetComment(ctx, commentID)
	if err != nil {
		if err == impart.ErrNotFound {
			return models.Comment{}, impart.NewError(err, "no comments found for post")
		} else {
			return models.Comment{}, impart.NewError(impart.ErrUnknown, "couldn't fetch comments")
		}
	}
	return models.CommentFromDBModel(dbComment), nil
}

func (s *service) NewComment(ctx context.Context, c models.Comment) (models.Comment, impart.Error) {
	var empty models.Comment
	var err error

	if len(strings.TrimSpace(c.Content.Markdown)) < 1 {
		return empty, impart.NewError(impart.ErrBadRequest, "post is less than 1 character1")
	}
	ctxUser := impart.GetCtxUser(ctx)
	newComment := &dbmodels.Comment{
		PostID:         c.PostID,
		ImpartWealthID: ctxUser.ImpartWealthID,
		CreatedAt:      impart.CurrentUTC(),
		Content:        c.Content.Markdown,
		LastReplyTS:    impart.CurrentUTC(),
		//ParentCommentID: c.p, //no threading for now
		UpVoteCount:   0,
		DownVoteCount: 0,
	}

	comment, err := s.commentData.NewComment(ctx, newComment)
	if err != nil {
		s.logger.Error("error creating comment", zap.Error(err), zap.Any("comment", comment))
		return models.Comment{}, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error creating NewComment for user %s", c.ImpartWealthID))
	}
	out := models.CommentFromDBModel(comment)
	dbPost, err := s.postData.GetPost(ctx, c.PostID)
	if err != nil {
		s.logger.Error("error getting post from newly created comment")
		return out, nil
	}
	previewEnd := maxNotificationLength
	if len(dbPost.Subject) < maxNotificationLength {
		previewEnd = len(dbPost.Subject)
	}

	previewText := dbPost.Subject[0:previewEnd]
	notificationData := impart.NotificationData{
		EventDatetime: impart.CurrentUTC(),
		PostID:        dbPost.PostID,
	}

	alert := impart.Alert{
		Title: aws.String(fmt.Sprintf("New Comment on Your Post: %s", dbPost.Subject)),
		Body:  aws.String(fmt.Sprintf("%s wrote %s", ctxUser.ScreenName, previewText)),
	}

	s.logger.Debug("sending notification", zap.Any("impartWealthId", dbPost.R.ImpartWealth.ImpartWealthID), zap.Any("alert", alert))

	go func() {
		if strings.TrimSpace(dbPost.R.ImpartWealth.ImpartWealthID) != "" {
			err = s.sendNotification(notificationData, alert, dbPost.R.ImpartWealth.ImpartWealthID)
			if err != nil {
				s.logger.Error("error attempting to send post reply ", zap.Error(err))
			}
		}
	}()

	return out, nil
}

func (s *service) EditComment(ctx context.Context, editedComment models.Comment) (models.Comment, impart.Error) {
	var empty models.Comment
	ctxUser := impart.GetCtxUser(ctx)
	existingComment, err := s.commentData.GetComment(ctx, editedComment.CommentID)
	if err != nil {
		s.logger.Error("error fetcing post trying to edit", zap.Error(err))
		return empty, impart.UnknownError
	}
	if !ctxUser.Admin && existingComment.ImpartWealthID != ctxUser.ImpartWealthID {
		return empty, impart.NewError(impart.ErrUnauthorized, "unable to edit a comment that's not yours")
	}
	existingComment.Content = editedComment.Content.Markdown
	c, err := s.commentData.EditComment(ctx, existingComment)
	if err != nil {
		return empty, impart.UnknownError
	}
	return models.CommentFromDBModel(c), nil
}

func (s *service) DeleteComment(ctx context.Context, commentID uint64) impart.Error {
	s.logger.Debug("received comment delete request",
		zap.Uint64("commentId", commentID))

	ctxUser := impart.GetCtxUser(ctx)
	existingComment, err := s.commentData.GetComment(ctx, commentID)
	if err != nil {
		return impart.NewError(err, "unable to locate existing comment to delete")
	}

	if !ctxUser.Admin && existingComment.ImpartWealthID != ctxUser.ImpartWealthID {
		return impart.NewError(impart.ErrUnauthorized, "unable to delete a comment that's not yours")
	}

	err = s.commentData.DeleteComment(ctx, commentID)
	if err != nil {
		if err == impart.ErrNotFound {
			return impart.NewError(err, "Comment not found")
		}
		return impart.NewError(err, "error deleting comment")
	}
	return nil
}

func (s *service) ReportComment(ctx context.Context, commentID uint64, reason string, remove bool) (models.PostCommentTrack, impart.Error) {
	var dbReason *string
	var empty models.PostCommentTrack
	if !remove && reason == "" {
		return empty, impart.NewError(impart.ErrBadRequest, "must provide a reason for reporting")
	}
	if reason != "" {
		dbReason = &reason
	}

	err := s.reactionData.ReportComment(ctx, commentID, dbReason, remove)
	if err != nil {
		s.logger.Error("couldn't report comment", zap.Error(err), zap.Uint64("commentId", commentID))
		switch err {
		case impart.ErrNoOp:
			return empty, impart.NewError(impart.ErrNoOp, "comment is already in the input reported state", impart.Report)
		case impart.ErrNotFound:
			return empty, impart.NewError(err, fmt.Sprintf("could not find comment %v to report", commentID), impart.Report)
		default:
			return empty, impart.UnknownError
		}
	}

	//send comment report notification
	err = s.SendCommentNotification(models.CommentNotificationInput{
		Ctx:             ctx,
		CommentID:       commentID,
		ActionType:      types.Report,
		ActionData:      reason,
		NotifyPostOwner: true,
	})

	if err != nil {
		s.logger.Error("error happened on notify reaction", zap.Error(err))
	}
	out, err := s.reactionData.GetUserTrack(ctx, data.ContentInput{
		Type: data.Comment,
		Id:   commentID,
	})
	if err != nil {
		s.logger.Error("couldn't get updated user track object", zap.Error(err))
		return empty, impart.UnknownError
	}
	return out, nil
}

/**
 * SendCommentReportNotification
 *
 * Send notification when a comment reported
 * Notifying to :
 *		post owner
 *		comment owner
 */
func (s *service) SendCommentNotification(input models.CommentNotificationInput) impart.Error {
	dbComment, err := s.commentData.GetComment(input.Ctx, input.CommentID)
	if err != nil {
		return impart.NewError(err, "unable to fetch comment for send notification")
	}
	// set post id in input
	input.PostID = dbComment.PostID

	notificationData := impart.NotificationData{
		EventDatetime: impart.CurrentUTC(),
		PostID:        0,
	}

	// generate notification context
	out, err := s.BuildNotificationData(input)
	if err != nil {
		return impart.NewError(err, "build comment notification params")
	}

	s.logger.Debug("sending comment notification", zap.Any("data", input), zap.Any("notificationData", out))

	// send to comment owner
	go func() {
		if strings.TrimSpace(dbComment.R.ImpartWealth.ImpartWealthID) != "" {
			err = s.sendNotification(notificationData, out.Alert, dbComment.R.ImpartWealth.ImpartWealthID)
			if err != nil {
				s.logger.Error("error attempting to send post comment notification ", zap.Error(err))
			}
		}
	}()

	// send to post owner
	if input.NotifyPostOwner {
		go func() {
			if strings.TrimSpace(out.PostOwnerWealthID) != "" {
				err = s.sendNotification(notificationData, out.PostOwnerAlert, out.PostOwnerWealthID)
				if err != nil {
					s.logger.Error("error attempting to send post comment notification post owner ", zap.Error(err))
				}
			}
		}()
	}

	return nil
}

//
// From here , all the notification action workflow
//
func (s *service) BuildNotificationData(input models.CommentNotificationInput) (models.CommentNotificationBuildDataOutput, error) {
	var previewText, postUserIWID string
	var alert, postOwnerAlert impart.Alert
	var err error
	var dbPost *dbmodels.Post

	ctxUser := impart.GetCtxUser(input.Ctx)

	// initialize dbPost
	if input.NotifyPostOwner {
		dbPost, err = s.postData.GetPost(input.Ctx, input.PostID)
		if err != nil {
			return models.CommentNotificationBuildDataOutput{}, impart.NewError(err, "unable to fetch comment post for send notification")
		}
	}

	switch input.ActionType {
	//in case of report
	case types.Report:
		previewEnd := maxNotificationLength
		if len(input.ActionData) < maxNotificationLength {
			previewEnd = len(input.ActionData)
		}
		previewText = input.ActionData[0:previewEnd]

		// make alert
		alert = impart.Alert{
			Title: aws.String("Comment has been reported"),
			Body: aws.String(
				fmt.Sprintf("%s wrote %s", ctxUser.ScreenName, previewText),
			),
		}

		// make post owner alert
		if input.NotifyPostOwner {
			postOwnerAlert = impart.Alert{
				Title: aws.String("A post comment has been reported"),
				Body: aws.String(
					fmt.Sprintf("%s on post %s wrote %s", ctxUser.ScreenName, dbPost.Subject, previewText),
				),
			}
			postUserIWID = dbPost.ImpartWealthID
		}
	case types.UpVote:
		// make alert
		alert = impart.Alert{
			Title: aws.String("Reacted on comment"),
			Body: aws.String(
				fmt.Sprintf("%s liked your comment", ctxUser.ScreenName),
			),
		}
		// make post owner alert
		if input.NotifyPostOwner {
			postOwnerAlert = impart.Alert{
				Title: aws.String("Reacted on post comment"),
				Body: aws.String(
					fmt.Sprintf("%s liked your %s post comment", ctxUser.ScreenName, dbPost.Subject),
				),
			}
			postUserIWID = dbPost.ImpartWealthID
		}

	case types.DownVote:
		// make alert
		alert = impart.Alert{
			Title: aws.String("Reacted on comment"),
			Body: aws.String(
				fmt.Sprintf("%s dis-liked your comment", ctxUser.ScreenName),
			),
		}
		// make post owner alert
		if input.NotifyPostOwner {
			postOwnerAlert = impart.Alert{
				Title: aws.String("Reacted on post comment"),
				Body: aws.String(
					fmt.Sprintf("%s dis-liked your %s post comment", ctxUser.ScreenName, dbPost.Subject),
				),
			}
			postUserIWID = dbPost.ImpartWealthID
		}
	default:
		err = impart.NewError(err, "invalid notify option")
	}

	return models.CommentNotificationBuildDataOutput{
		Alert:             alert,
		PostOwnerAlert:    postOwnerAlert,
		PostOwnerWealthID: postUserIWID,
	}, err
}
