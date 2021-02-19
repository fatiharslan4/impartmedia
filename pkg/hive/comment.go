package hive

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	hive_data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

func (s *service) GetComments(hiveID, postID string, limit int64, nextPage *models.NextPage, authenticationID string) (models.Comments, *models.NextPage, impart.Error) {
	p, impartErr := s.validateHiveAccess(hiveID, authenticationID)
	if impartErr != nil {
		return models.Comments{}, nil, impartErr
	}

	comments, nextPage, err := s.commentData.GetComments(postID, limit, nextPage)
	if err != nil {
		s.logger.Error("error getting comment", zap.Error(err),
			zap.String("hiveId", hiveID),
			zap.String("postId", postID),
			zap.Int64("limit", limit),
			zap.Any("offset", nextPage),
			zap.String("authenticationId", authenticationID))
		return models.Comments{}, nextPage, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to returns comments for post %s", postID))
	}

	comments, err = s.addCommentTracks(p.ImpartWealthID, hiveID, postID, comments)
	if err != nil && err != impart.ErrNotFound {
		s.logger.Error("error getting comment tracks", zap.Error(err),
			zap.String("hiveID", hiveID), zap.String("postID", postID))
		return models.Comments{}, nextPage, impart.NewError(err, "error retrieving comment tracks")
	}

	comments.SortAscending()
	return comments, nextPage, nil
}

func (s *service) addCommentTracks(impartWealthID, hiveID, postID string, comments models.Comments) (models.Comments, error) {

	if len(comments) > 0 {
		commentIDs := comments.ContentIDs()
		s.logger.Debug("getting comment tracking data", zap.Strings("contentIDs", commentIDs))

		commentTracks, err := s.trackStore.GetUserTrackForContent(impartWealthID, comments.ContentIDs())
		if err != nil {
			if err != impart.ErrNotFound {
				s.logger.Error("error getting comment tracks", zap.Error(err),
					zap.String("hiveID", hiveID), zap.String("postID", postID))
				return comments, err
			}
			s.logger.Debug("no comment tracks found")

		}
		s.logger.Debug("retrieved comment tracks", zap.Any("tracks", commentTracks))
		comments.AppendContentTracks(commentTracks)
	}
	return comments, nil
}

//func (s *service) GetCommentsByImpartWealthID(impartWealthID string, limit int64, offset time.Time, authenticationID string) (models.Comments, impart.Error) {
//	comments, err := s.commentData.GetCommentsByImpartWealthID(impartWealthID, limit, offset)
//	if err != nil {
//		return models.Comments{}, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to returns comments for impartWealthID %s", impartWealthID))
//	}
//
//	return comments, nil
//}

func (s *service) GetComment(hiveID, postID, commentID string, consistentRead bool, authenticationID string) (models.Comment, impart.Error) {
	p, impartErr := s.validateHiveAccess(hiveID, authenticationID)
	if impartErr != nil {
		return models.Comment{}, impartErr
	}

	comment, err := s.commentData.GetComment(postID, commentID, false)
	if err != nil {
		s.logger.Error("error getting comment", zap.Error(err),
			zap.String("hiveId", hiveID),
			zap.String("postId", postID),
			zap.String("commentId", commentID),
			zap.String("authenticationId", authenticationID))
		return models.Comment{}, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to get comment %s for postID %s", commentID, postID))
	}

	comment.PostCommentTrack, err = s.trackStore.GetUserTrack(p.ImpartWealthID, comment.CommentID, false)
	if err != nil && err != impart.ErrNotFound {
		return models.Comment{}, impart.NewError(err, "error getting tracked data for comment")
	}

	return comment, nil
}

func (s *service) NewComment(c models.Comment, authenticationID string) (models.Comment, impart.Error) {
	var profile models.Profile
	var out models.Comment
	var post models.Post
	var impartErr impart.Error
	var err error

	if len(strings.TrimSpace(c.Content.Markdown)) < 1 {
		return out, impart.NewError(impart.ErrBadRequest, "post is less than 1 character1")
	}

	// Check validations asynchronously
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		profile, impartErr = s.validateHiveAccess(c.HiveID, authenticationID)
		if impartErr != nil {
			s.logger.Error("cannot validate the user is able to post to this hive", zap.Any("impartError", impartErr))
		}
	}()

	go func() {
		defer wg.Done()
		post, err = s.postData.GetPost(c.HiveID, c.PostID, false)
		if err != nil {
			if err == impart.ErrNotFound {
				impartErr = impart.NewError(err, "post does not exist")
			}
			impartErr = impart.NewError(err, "error getting existing post for comment")
		}
	}()
	//Wait for both validations
	wg.Wait()

	if impartErr != nil {
		return out, impartErr
	}

	c.ImpartWealthID = profile.ImpartWealthID
	c.ScreenName = profile.ScreenName
	c.CommentID = ksuid.New().String()
	c.CommentDatetime = impart.CurrentUTC()
	c.Edits = models.Edits{}
	c.UpVotes = 0
	c.DownVotes = 0

	s.logger.Debug("creating comment in data store", zap.Any("comment", c))

	comment, err := s.commentData.NewComment(c)
	if err != nil {
		s.logger.Error("error creating comment", zap.Error(err), zap.Any("comment", comment))
		return models.Comment{}, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error creating NewComment for user %s", c.ImpartWealthID))
	}

	//These waitgroup funcs log errors they encounter, but do not return any errors to the caller as they are non essential to the request.
	wg = sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		if err = s.postData.UpdateTimestampLater(post.HiveID, post.PostID, hive_data.LastCommentDatetimeColumnName, c.CommentDatetime); err != nil {
			s.logger.Error("Error updating timestamp", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		if err = s.postData.IncrementDecrementPost(post.HiveID, post.PostID, hive_data.CommentCountColumnName, false); err != nil {
			s.logger.Error("Error updating timestamp", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		posterProfile, err := s.profileData.GetProfile(post.ImpartWealthID, false)
		if err != nil {
			s.logger.Error("error getting posters profile", zap.Error(err))
			return
		}
		np := posterProfile.NotificationProfile

		previewEnd := maxNotificationLength
		if len(post.Subject) < maxNotificationLength {
			previewEnd = len(post.Subject)
		}

		previewText := post.Subject[0:previewEnd]
		notificationData := impart.NotificationData{
			EventDatetime: impart.CurrentUTC(),
			PostID:        c.PostID,
		}

		alert := impart.Alert{
			Title: aws.String(fmt.Sprintf("New Comment on Your Post: %s", post.Subject)),
			Body:  aws.String(fmt.Sprintf("%s wrote %s", comment.ScreenName, previewText)),
		}

		s.logger.Debug("sending notification", zap.Any("notificationProfile", np), zap.Any("alert", alert))

		if strings.TrimSpace(np.DeviceToken) != "" {
			err = s.sendNotification(notificationData, alert, posterProfile)
			if err != nil {
				s.logger.Error("error attempting to send post reply ", zap.Error(err))
			}
		}
	}()

	//We don't wait here cause we need the outputs above, we wait because lambda runtime can shut this down before logging
	// any output or actually making these requests.
	wg.Wait()
	return comment, nil
}

func (s *service) EditComment(editedComment models.Comment, authenticationID string) (models.Comment, impart.Error) {
	var out models.Comment

	existingComment, err := s.commentData.GetComment(editedComment.PostID, editedComment.CommentID, true)
	if err != nil {
		return out, impart.NewError(err, "unable to locate existing comment to edit")
	}

	profile, impartErr := s.selfOrAdmin(existingComment.HiveID, existingComment.ImpartWealthID, authenticationID)
	if impartErr != nil {
		s.logger.Error("user is not authorized to edit this comment",
			zap.Any("comment", editedComment), zap.String("authenticationId", authenticationID))
		return models.Comment{}, impartErr
	}

	if existingComment.Content.Markdown == editedComment.Content.Markdown {
		return out, impart.NewError(impart.ErrBadRequest, "content has not changed")
	}

	newEdit := models.Edit{
		Datetime:       impart.CurrentUTC(),
		ImpartWealthID: profile.ImpartWealthID,
		ScreenName:     profile.ScreenName,
	}

	existingComment.Edits = append(existingComment.Edits, newEdit)
	existingComment.Content = editedComment.Content

	s.logger.Debug("Editing post", zap.Any("edit", newEdit), zap.Any("updated post", existingComment))

	out, err = s.commentData.EditComment(existingComment)
	if err != nil {
		s.logger.Error("error editing comment", zap.Error(err), zap.Any("comment", existingComment))
		return models.Comment{}, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to get edit comment %s for postID %s", editedComment.CommentID, editedComment.PostID))
	}

	return out, nil
}

func (s *service) DeleteComment(postID, commentID, authenticationID string) impart.Error {
	s.logger.Debug("received comment delete request", zap.String("postId", postID),
		zap.String("commentId", commentID),
		zap.String("authenticationId", authenticationID))

	existingComment, err := s.commentData.GetComment(postID, commentID, true)
	if err != nil {
		return impart.NewError(err, "unable to locate existing comment to delete")
	}

	_, impartErr := s.selfOrAdmin(existingComment.HiveID, existingComment.ImpartWealthID, authenticationID)
	if impartErr != nil {
		s.logger.Error("user is not authorized to edit this comment",
			zap.Any("comment", existingComment), zap.String("authenticationId", authenticationID))
		return impartErr
	}

	err = s.commentData.DeleteComment(postID, commentID)
	if err != nil {
		if err == impart.ErrNotFound {
			return impart.NewError(err, "Comment not found")
		}
		return impart.NewError(err, "error deleting comment")
	}
	err = s.postData.IncrementDecrementPost(existingComment.HiveID, postID, hive_data.CommentCountColumnName, true)
	if err != nil {
		s.logger.Error("error decrementing post comments ")
	}
	return nil
}
