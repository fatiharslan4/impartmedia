package hive

import (
	"context"
	"fmt"
	"strings"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
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
		return empty, impart.NewError(impart.ErrBadRequest, "post is less than 1 character1", impart.Content)
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
		return empty, impart.NewError(impart.ErrUnauthorized, "unable to edit a comment that's not yours", impart.ImpartWealthID)
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
			return empty, impart.NewError(impart.ErrNoOp, "comment is already in the input reported state")
		case impart.ErrNotFound:
			return empty, impart.NewError(err, fmt.Sprintf("could not find comment %v to report", commentID))
		default:
			return empty, impart.UnknownError
		}
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
