package hive

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/impartwealthapp/backend/pkg/impart"
	"go.uber.org/zap"
)

const adminPostNotification = "Your Hive is Buzzing"

// This is a lot of branches; should probably be broken up.
func (s *service) PinPost(ctx context.Context, hiveID, postID uint64, pin bool) impart.Error {
	ctxUser := impart.GetCtxUser(ctx)

	if !ctxUser.Admin {
		return impart.NewError(impart.ErrUnauthorized, "cannot pin a post unless you are a hive admin")
	}
	fmt.Println("the post pin are")
	err := s.hiveData.PinPost(ctx, hiveID, postID, pin)
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
	fmt.Println("the pin", pin, dbHive.PinnedPostID.Uint64, dbPost.PostID)
	if pin && dbHive.PinnedPostID.Uint64 == dbPost.PostID {
		pushNotification := impart.Alert{
			Title: aws.String(adminPostNotification),
			Body:  aws.String(dbPost.Subject),
		}

		additionalData := impart.NotificationData{
			EventDatetime: impart.CurrentUTC(),
			PostID:        dbPost.PostID,
		}

		fmt.Println("the topic are", dbHive.NotificationTopicArn.String)

		err = s.notificationService.NotifyTopic(ctx, additionalData, pushNotification, dbHive.NotificationTopicArn.String)
		if err != nil {
			s.logger.Error("error sending notification to topic", zap.Error(err))
		}

	}

	return nil
}
