package hive

import (
	"fmt"
	"strings"

	data "github.com/impartwealthapp/backend/pkg/data/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

// VoteInput is the input to register an upvote or downvote on a comment or post.
type VoteInput struct {
	HiveID, PostID, CommentID string
	Upvote                    bool
	Increment                 bool
}

func (s *service) Votes(v VoteInput, authID string) (models.PostCommentTrack, impart.Error) {
	var out models.PostCommentTrack
	s.logger.Debug("received vote request", zap.Any("input", v))

	p, impartErr := s.validateHiveAccess(v.HiveID, authID)
	if impartErr != nil {
		return out, impartErr
	}

	var contentID string
	if len(strings.TrimSpace(v.CommentID)) == 27 {
		contentID = v.CommentID
	} else {
		contentID = v.PostID
	}

	var err error
	if v.Upvote {
		if v.Increment {
			err = s.trackStore.AddUpVote(p.ImpartWealthID, contentID, v.HiveID, v.PostID)
		} else {
			err = s.trackStore.TakeUpVote(p.ImpartWealthID, contentID, v.HiveID, v.PostID)
		}
	} else { //Downvote
		if v.Increment {
			err = s.trackStore.AddDownVote(p.ImpartWealthID, contentID, v.HiveID, v.PostID)
		} else {
			err = s.trackStore.TakeDownVote(p.ImpartWealthID, contentID, v.HiveID, v.PostID)
		}
	}

	if err != nil && err != impart.ErrNoOp {
		s.logger.Error("error updating track store", zap.Error(err), zap.Any("vote", v), zap.Any("profile", p))
		return out, impart.NewError(err, "error updating customers vote tracking")
	}

	out, err = s.trackStore.GetUserTrack(p.ImpartWealthID, contentID, true)
	if err != nil {
		s.logger.Error("error getting updated tracked item track store", zap.Error(err), zap.Any("vote", v), zap.Any("profile", p))
		return out, impart.NewError(err, "unable to retrieve recently tracked content")
	}

	return out, nil
}

func (s *service) CommentCount(hiveID, postID string, subtract bool, authenticationID string) impart.Error {
	_, impartErr := s.validateHiveAccess(hiveID, authenticationID)
	if impartErr != nil {
		return impartErr
	}

	if err := s.postData.IncrementDecrementPost(hiveID, postID, data.CommentCountColumnName, subtract); err != nil {
		return impart.NewError(err, "error incrementing commentCount")
	}
	return nil
}

func (s *service) Logger() *zap.Logger {
	return s.logger
}

func (s *service) sendNotification(data impart.NotificationData, alert impart.Alert, profile models.Profile) error {
	if profile.NotificationProfile.DeviceToken != "" {
		sentARN, err := s.notificationService.NotifyAppleDevice(data, alert, profile.NotificationProfile.DeviceToken, profile.NotificationProfile.AWSPlatformEndpointARN)
		if err != nil {
			return err
		}
		if sentARN != profile.NotificationProfile.AWSPlatformEndpointARN {
			//update the users AWS ARN for notifications
			profile.NotificationProfile.AWSPlatformEndpointARN = sentARN
			err := s.profileData.UpdateProfileProperty(profile.ImpartWealthID, "notificationProfile.awsPlatformEndpointARN", profile.NotificationProfile.AWSPlatformEndpointARN)
			if err != nil {
				return err
			}
		}
	} else {
		s.logger.Debug("profile has no notification profile", zap.Any("profile", profile))
	}
	return nil
}

func (s *service) selfOrAdmin(hiveID, originalContentImpartWealthID, authenticationID string) (models.Profile, impart.Error) {

	profile, err := s.profileData.GetProfileFromAuthId(authenticationID, false)
	if err != nil {
		s.logger.Error("Profile Data not retrieved", zap.String("authenticationId", authenticationID))
		return models.Profile{}, impart.NewError(err, "unable to map authenticationId to an impart wealth user")
	}

	if profile.ImpartWealthID == originalContentImpartWealthID || profile.Attributes.Admin {
		s.logger.Debug("impartWealthId of authenticated user matches the impartWealthId of the original content user,"+
			"or user is admin",
			zap.String("original", originalContentImpartWealthID),
			zap.String("authenticatedUser", profile.ImpartWealthID),
			zap.Bool("isAdmin", profile.Attributes.Admin))
		return profile, nil
	}

	hive, err := s.hiveData.GetHive(hiveID, false)
	if err != nil {
		s.logger.Error("error getting hive data", zap.String("hiveId", hiveID))
		return models.Profile{}, impart.NewError(err, "unable to validate admin privileges for editing the post.")
	}

	var isAdmin bool
	for _, a := range hive.Administrators {
		if a.ImpartWealthID == profile.ImpartWealthID {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		s.logger.Warn("user is not a hive admin, and is attempting to modify something not their own. ")
		return models.Profile{}, impart.NewError(impart.ErrUnauthorized, "user not authorized to take this action on the resource.")
	}

	return profile, nil
}

func (s *service) validateHiveAccess(hiveID, authenticationID string) (models.Profile, impart.Error) {
	if _, err := s.hiveData.GetHive(hiveID, false); err != nil {
		return models.Profile{}, impart.NewError(err, fmt.Sprintf("unable to retreive hive %s", hiveID))
	}

	profile, err := s.profileData.GetProfileFromAuthId(authenticationID, false)
	if err != nil {
		return models.Profile{}, impart.NewError(err, "unable to map authenticationId to an impart wealth user")
	}

	if !isMember(hiveID, profile) && !profile.Attributes.Admin {
		return models.Profile{}, impart.NewError(impart.ErrUnauthorized, fmt.Sprintf("user is not a member of "+
			"hive '%s'; denied.", hiveID))
	}

	return profile, nil
}

func isMember(hiveID string, profile models.Profile) bool {
	for _, m := range profile.Attributes.HiveMemberships {
		if m.HiveID == hiveID {

			return true
		}
	}
	return false
}

func (s *service) GetHive(authID, hiveID string) (models.Hive, impart.Error) {
	if _, err := s.validateHiveAccess(hiveID, authID); err != nil {
		return models.Hive{}, err
	}

	hive, err := s.hiveData.GetHive(hiveID, false)
	if err != nil {
		s.logger.Error("error getting hive", zap.Error(err))
		if err == impart.ErrNotFound {
			return hive, impart.NewError(err, fmt.Sprintf("hive %s not found", hiveID))
		}
		return hive, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to retrieve hive %s", hiveID))
	}
	return hive, nil
}

// If the auth user is an admin, then return all hives.  Otherwise only return hives the user is a member of.
func (s *service) GetHives(authID string) (models.Hives, impart.Error) {
	var err error
	var profile models.Profile
	var hives models.Hives
	profile, err = s.profileData.GetProfileFromAuthId(authID, false)
	if err != nil {
		return hives, impart.NewError(err, "error looking up profile using authenticationId")
	}

	if profile.Attributes.Admin {
		hives, err := s.hiveData.GetHives()
		if err != nil {
			return hives, impart.NewError(impart.ErrUnknown, "error when attempting to retrieve hives")
		}
		return hives, nil
	}
	hives = make(models.Hives, 0, 0)

	for _, hm := range profile.Attributes.HiveMemberships {
		h, err := s.hiveData.GetHive(hm.HiveID, false)
		if err != nil {
			return models.Hives{}, impart.NewError(err, "could not retrieve hive id")
		}
		hives = append(hives, h)
	}

	return hives, nil
}

func (s *service) CreateHive(authID string, hive models.Hive) (models.Hive, impart.Error) {
	var err error
	var profile models.Profile
	profile, err = s.profileData.GetProfileFromAuthId(authID, false)
	if err != nil {
		return models.Hive{}, impart.NewError(err, "error retrieving profile for hives")
	}

	if !profile.Attributes.Admin {
		return models.Hive{}, impart.NewError(impart.ErrUnauthorized, "non-admin users cannot create hives.")
	}

	hive.HiveID = ksuid.New().String()

	hive, err = s.hiveData.NewHive(hive)
	if err != nil {
		return hive, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to create hive %s", hive.HiveName))
	}
	return hive, nil
}

func (s *service) EditHive(authID string, hive models.Hive) (models.Hive, impart.Error) {
	var err error
	var profile models.Profile
	profile, err = s.profileData.GetProfileFromAuthId(authID, false)
	if err != nil {
		return models.Hive{}, impart.NewError(err, "error retrieving profile for hives")
	}

	if !profile.Attributes.Admin {
		return models.Hive{}, impart.NewError(impart.ErrUnauthorized, "non-admin users cannot create hives.")
	}

	hive, err = s.hiveData.EditHive(hive)
	if err != nil {
		return hive, impart.NewError(impart.ErrUnknown, fmt.Sprintf("error when attempting to edit hive %s", hive.HiveName))
	}
	return hive, nil
}
