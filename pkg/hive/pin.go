package hive

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/impartwealthapp/backend/pkg/impart"
	"go.uber.org/zap"
)

const adminPostNotification = "Your Hive is Buzzing"
const title = "New post"
const body = "Admin added a post on Your Hive"

// This is a lot of branches; should probably be broken up.
func (s *service) PinPost(ctx context.Context, hiveID, postID uint64, pin bool, isAdminActivity bool) impart.Error {
	// ctxUser := impart.GetCtxUser(ctx)

	if !isAdminActivity {
		return impart.NewError(impart.ErrUnauthorized, "cannot pin a post unless you are a hive admin")
	}
	err := s.hiveData.PinPost(ctx, hiveID, postID, pin, isAdminActivity)
	if err != nil {
		if err == impart.ErrNoOp {
			return nil
		}
		s.logger.Error("error pining post", zap.Error(err))
		return impart.NewError(impart.ErrUnknown, "unable to pin post")
	}
	dbHive, err := s.hiveData.GetHive(ctx, hiveID)
	if err != nil {
		s.logger.Error("error pining post", zap.Error(err))
		return impart.NewError(impart.ErrUnknown, "unable to pin post")
	}
	dbPost, err := s.postData.GetPost(ctx, postID)
	if err != nil {
		s.logger.Error("error pining post", zap.Error(err))
		return impart.NewError(impart.ErrUnknown, "unable to pin post")
	}
	if pin && dbHive.PinnedPostID.Uint64 == dbPost.PostID {
		pushNotification := impart.Alert{
			Title: aws.String(title),
			Body:  aws.String(body),
		}

		additionalData := impart.NotificationData{
			EventDatetime: impart.CurrentUTC(),
			PostID:        dbPost.PostID,
		}

		err = s.notificationService.NotifyTopic(ctx, additionalData, pushNotification, dbHive.NotificationTopicArn.String)
		if err != nil {
			s.logger.Error("error sending notification to topic", zap.Error(err))
		}

	}

	return nil
}

// This is a lot of branches; should probably be broken up.
func (s *service) PinPostForBulkPostAction(ctx context.Context, postHive map[uint64]uint64, pin bool, isAdminActivity bool) impart.Error {

	if !isAdminActivity {
		return impart.NewError(impart.ErrUnauthorized, "cannot pin a post unless you are a hive admin")
	}
	err := s.hiveData.PinPostForBulkPostAction(ctx, postHive, pin, isAdminActivity)
	if err != nil {
		if err == impart.ErrNoOp {
			return nil
		}
		s.logger.Error("error pining post", zap.Error(err))
		return impart.NewError(impart.ErrUnknown, "unable to pin post")
	}
	// dbHive, err := s.hiveData.GetHive(ctx, 1)
	// if err != nil {
	// 	s.logger.Error("error pining post", zap.Error(err))
	// 	return impart.NewError(impart.ErrUnknown, "unable to pin post")
	// }
	// dbPost, err := s.postData.GetPost(ctx, 1)
	// if err != nil {
	// 	s.logger.Error("error pining post", zap.Error(err))
	// 	return impart.NewError(impart.ErrUnknown, "unable to pin post")
	// }
	// if pin && dbHive.PinnedPostID.Uint64 == dbPost.PostID {
	// 	pushNotification := impart.Alert{
	// 		Title: aws.String(title),
	// 		Body:  aws.String(body),
	// 	}

	// 	additionalData := impart.NotificationData{
	// 		EventDatetime: impart.CurrentUTC(),
	// 		PostID:        dbPost.PostID,
	// 	}

	// 	err = s.notificationService.NotifyTopic(ctx, additionalData, pushNotification, dbHive.NotificationTopicArn.String)
	// 	if err != nil {
	// 		s.logger.Error("error sending notification to topic", zap.Error(err))
	// 	}

	// }

	return nil
}
