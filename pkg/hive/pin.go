package hive

import (
	"github.com/aws/aws-sdk-go/aws"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

const adminPostNotification = "Your Hive is Buzzing"

// This is a lot of branches; should probably be broken up.
func (s *service) PinPost(hiveID, postID, authenticationID string, pin bool) impart.Error {
	var h models.Hive
	var p models.Post
	var impartErr impart.Error
	var err error

	profile, err := s.profileData.GetProfileFromAuthId(authenticationID, false)
	if err != nil {
		return impart.NewError(err, "unable to map authenticationId to an impart wealth user")
	}

	if !profile.Attributes.Admin {
		return impart.NewError(impart.ErrUnauthorized, "cannot pin a post unless you are a hive admin")
	}

	if p, impartErr = s.GetPost(hiveID, postID, false, authenticationID); impartErr != nil {
		return impartErr
	}

	if h, impartErr = s.GetHive(authenticationID, hiveID); impartErr != nil {
		return impartErr
	}

	if pin {
		if h.PinnedPostID != p.PostID {
			if h.PinnedPostID != "" {
				if err = s.postData.SetPinStatus(hiveID, h.PinnedPostID, false); err != nil {
					return impart.NewError(err, "error un-setting pinned post status")
				}
			}

			if err = s.hiveData.PinPost(hiveID, postID); err != nil {
				return impart.NewError(err, "error setting postID on hive")
			}

			pushNotification := impart.Alert{
				Title: aws.String(adminPostNotification),
				Body:  aws.String(p.Subject),
			}

			additionalData := impart.NotificationData{
				EventDatetime: impart.CurrentUTC(),
				PostID:        p.PostID,
			}

			err = s.notificationService.NotifyTopic(additionalData, pushNotification, h.PinnedPostNotificationTopicARN)
			if err != nil {
				s.logger.Error("error sending notification to topic", zap.Error(err))
			}
		}

		if !p.IsPinnedPost {
			if err = s.postData.SetPinStatus(hiveID, postID, true); err != nil {
				return impart.NewError(err, "error setting post to pinned")
			}
		}
	} else {
		// unpin or not pinned
		if h.PinnedPostID != "" {
			if err = s.hiveData.PinPost(hiveID, ""); err != nil {
				return impart.NewError(err, "error un-setting hives pinned postID")
			}
		}

		if p.IsPinnedPost {
			if err = s.postData.SetPinStatus(hiveID, postID, false); err != nil {
				return impart.NewError(err, "error un-setting post to unpinned")
			}
		}
	}

	return nil
}
