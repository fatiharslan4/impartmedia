package impart

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

const notificationProfilePropertyPath = "notificationProfile.awsPlatformEndpointARN"

type NotificationService interface {
	Notify(ctx context.Context, data NotificationData, alert Alert, impartWealthID string) error
	NotifyAppleDevice(ctx context.Context, data NotificationData, alert Alert, deviceToken, platformEndpointARN string) (sentPlatformEndpointARN string, err error)
	NotifyTopic(ctx context.Context, data NotificationData, alert Alert, topicARN string) error
	// subscribtion to topic methods
	SubscribeTopic(ctx context.Context, impartWealthID, topicARN, platformEndpointARN string) error
	UnsubscribeTopic(ctx context.Context, impartWealthID string, subscriptionARN string) (err error)
	UnsubscribeAll(ctx context.Context, impartWealthID string) error
	UnsubscribeTopicForDevice(ctx context.Context, impartWealthID, topicARN, platformEndpointARN string) error

	// SyncTokenEndpoint is meant to be called when a profiles deviceToken has been updated - this will ensure that the platformApplication
	// has the right device token, and the endpoint is enabled.
	SyncTokenEndpoint(ctx context.Context, deviceToken, platformEndpointARN string) (string, error)
}

type NotificationData struct {
	EventDatetime time.Time `json:"eventDatetime"`
	PostID        uint64    `json:"postId,omitempty"`
}

type noopNotificationService struct {
}

func (n noopNotificationService) Notify(ctx context.Context, data NotificationData, alert Alert, impartWealthID string) error {
	return nil
}

func (n noopNotificationService) UnsubscribeAll(ctx context.Context, impartWealthID string) error {
	return nil
}

func (n noopNotificationService) SyncTokenEndpoint(ctx context.Context, deviceToken, platformEndpointARN string) (string, error) {
	return "", nil
}

func (n noopNotificationService) NotifyTopic(ctx context.Context, data NotificationData, alert Alert, topicARN string) error {
	return nil
}

func (n noopNotificationService) SubscribeTopic(ctx context.Context, platformApplicationARN, topicARN, platformEndpointARN string) error {
	return nil
}

func (n noopNotificationService) UnsubscribeTopic(ctx context.Context, impartWealthID string, subscriptionARN string) error {
	return nil
}

func (n noopNotificationService) NotifyAppleDevice(ctx context.Context, data NotificationData, alert Alert, deviceToken, platformEndpointARN string) (string, error) {
	return "", nil
}

func (n noopNotificationService) UnsubscribeTopicForDevice(ctx context.Context, impartWealthID, topicARN, platformEndpointARN string) error {
	return nil
}

func NewNoopNotificationService() NotificationService {
	return &noopNotificationService{}
}

type snsAppleNotificationService struct {
	stage string
	*sns.SNS
	*zap.Logger
	platformApplicationARN string
	db                     *sql.DB
}

type Alert struct {
	Title    *string `json:"title"`
	SubTitle *string `json:"subtitle"`
	Body     *string `json:"body"`
}

type APNSMessage struct {
	// The message to display
	Alert Alert `json:"alert,omitempty"`
	// The sound to make
	Sound *string `json:"sound,omitempty"`
	// Any custom data to be included in the alert
	Data interface{} `json:"custom_data"`
	// The badge to show
	Badge *int `json:"badge,omitempty"`
}

type apnsMessageWrapper struct {
	// The top level apnsData set
	APNSData APNSMessage `json:"aps"`
}

type awsSNSMessage struct {
	APNS        string `json:"APNS"`
	APNSSandbox string `json:"APNS_SANDBOX"`
	Default     string `json:"default"`
	GCM         string `json:"GCM"`
}

func NewImpartNotificationService(db *sql.DB, stage, region, platformApplicationARN string, logger *zap.Logger) NotificationService {

	//SNS not available in us-east-2
	if strings.EqualFold(region, "us-east-2") {
		region = "us-east-1"
	}
	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String(region),
		HTTPClient: NewHttpClient(10 * time.Second),
	})
	if err != nil {
		logger.Fatal("unable to create aws session", zap.Error(err))
	}

	snsAppleNotificationService := &snsAppleNotificationService{
		stage:                  stage,
		Logger:                 logger,
		SNS:                    sns.New(sess),
		platformApplicationARN: platformApplicationARN,
		db:                     db,
	}

	logger.Debug("created new NotificationService",
		zap.String("stage", stage),
		zap.String("arn", platformApplicationARN))

	return snsAppleNotificationService
}

func (ns *snsAppleNotificationService) NotifyTopic(ctx context.Context, data NotificationData, alert Alert, topicARN string) error {
	var b []byte
	var err error

	if strings.TrimSpace(topicARN) == "" {
		return nil
	}

	ns.Logger.Debug("sending push notification",
		zap.Any("data", data),
		zap.Any("msg", alert),
		zap.String("platformEndpoint", topicARN),
		zap.String("arn", ns.platformApplicationARN))

	if b, err = json.Marshal(apnsMessageWrapper{
		APNSData: APNSMessage{
			Alert: alert,
			Sound: aws.String("default"),
			Data:  data,
			Badge: aws.Int(0),
		},
	}); err != nil {
		return err
	}

	msg := awsSNSMessage{Default: *alert.Body}

	msg.APNS = string(b)
	msg.APNSSandbox = string(b)

	if b, err = json.Marshal(msg); err != nil {
		return err
	}

	input := &sns.PublishInput{
		Message:          aws.String(string(b)),
		MessageStructure: aws.String("json"),
		TopicArn:         aws.String(topicARN),
	}
	// print()
	_, err = ns.Publish(input)
	return err
}

//
// Notification only send to active devices of user
//
// only fectch 5 active devices of user
func (ns *snsAppleNotificationService) Notify(ctx context.Context, data NotificationData, alert Alert, impartWealthID string) error {
	activeDevices, err := dbmodels.NotificationDeviceMappings(
		dbmodels.NotificationDeviceMappingWhere.ImpartWealthID.EQ(impartWealthID),
		dbmodels.NotificationDeviceMappingWhere.NotifyStatus.EQ(true),
		qm.Offset(0),
		qm.Limit(5),
		qm.OrderBy("map_id desc"),
		qm.Load(dbmodels.NotificationDeviceMappingRels.UserDevice),
		qm.Load(dbmodels.NotificationDeviceMappingRels.ImpartWealth),
	).All(ctx, ns.db)

	if err != nil {
		return fmt.Errorf("unable to fetch user active devices %v", err)
	}
	// if not active devices found
	if len(activeDevices) <= 0 {
		ns.Logger.Info("no active devices found for user",
			zap.Any("msg", alert),
			zap.String("impartWealthID", impartWealthID),
		)
	}

	// loop through the active devices and send notification
	for _, u := range activeDevices {
		if u.NotifyArn == "" {
			return fmt.Errorf("empty device token found for user %v", impartWealthID)
		}
		if u.R.UserDevice == nil {
			return fmt.Errorf("unable to find user device information")
		}
		// user device
		userDevice := u.R.UserDevice
		// var notificationStatus bool

		_, err := ns.NotifyAppleDevice(ctx, data, alert, userDevice.DeviceToken, u.NotifyArn)
		if err != nil {
			ns.Logger.Error("unable to notify to the device", zap.Any("device", userDevice))
			continue
		}

		// if snsAppARN != u.AwsSNSAppArn {
		// 	u.AwsSNSAppArn = snsAppARN
		// 	if _, err := u.Update(ctx, ns.db, boil.Whitelist(dbmodels.UserColumns.AwsSNSAppArn)); err != nil {
		// 		ns.Logger.Error("unable to update sns app arn")
		// 	}
		// }
	}
	return nil
}

func (ns *snsAppleNotificationService) NotifyAppleDevice(ctx context.Context, data NotificationData, alert Alert, deviceToken, platformEndpointARN string) (string, error) {
	var b []byte
	var err error

	platformEndpointARN, err = ns.SyncTokenEndpoint(ctx, deviceToken, platformEndpointARN)
	if err != nil {
		return "", err
	}

	ns.Logger.Debug("sending push notification",
		zap.Any("msg", alert),
		zap.String("deviceToken", deviceToken),
		zap.String("platformEndpoint", platformEndpointARN),
		zap.String("arn", ns.platformApplicationARN))

	if b, err = json.Marshal(apnsMessageWrapper{
		APNSData: APNSMessage{
			Alert: alert,
			Sound: aws.String("default"),
			Data:  data,
			Badge: aws.Int(0),
		},
	}); err != nil {
		return "", err
	}

	msg := awsSNSMessage{Default: *alert.Body}

	msg.APNS = string(b)
	msg.APNSSandbox = string(b)

	if b, err = json.Marshal(msg); err != nil {
		return "", err
	}

	input := &sns.PublishInput{
		Message:          aws.String(string(b)),
		MessageStructure: aws.String("json"),
		TargetArn:        aws.String(platformEndpointARN),
	}
	_, err = ns.Publish(input)
	return platformEndpointARN, err
}

func (ns *snsAppleNotificationService) SyncTokenEndpoint(ctx context.Context, deviceToken, platformEndpointARN string) (string, error) {
	var err error
	// No stored endpoint ARN
	if strings.TrimSpace(platformEndpointARN) == "" {
		ns.Logger.Debug("didn't receive a stored endpoint - attempting to create one.")
		platformEndpointARN, err = ns.createEndpoint(ctx, deviceToken)
		if err != nil {
			ns.Logger.Error("error creating endpoint", zap.Error(err))
			return "", err
		}

	}

	// Check existing endpoint
	resp, err := ns.GetEndpointAttributes(&sns.GetEndpointAttributesInput{
		EndpointArn: aws.String(platformEndpointARN),
	})

	if err != nil {
		ns.Debug("endpoint received an error", zap.Error(err), zap.Any("getEndpointResponse", resp))

		if awsErr, ok := err.(awserr.Error); !ok ||
			awsErr.Code() != sns.ErrCodeNotFoundException {
			ns.Logger.Error("error getting endpoint attributes", zap.Error(err))
			return "", err
		} else {
			//It is a not found exception, so just create it
			ns.Debug("endpoint not found, creating a new fresh endpoint")
			return ns.createEndpoint(ctx, deviceToken)
		}
	}
	// the endpoint was found, ensure it is enabled and has the right deviceToken
	endpointEnabled, err := strconv.ParseBool(*resp.Attributes["Enabled"])
	if err != nil {
		ns.Logger.Error("error parsing Enabled Attribute of response", zap.Error(err))
		return "", err
	}

	//if the endpoint was found, but is a different token or is disabled, update the properties to re-enable it.
	if *resp.Attributes["Token"] != deviceToken || !endpointEnabled {
		ns.Debug("Endpoint was disabled or device tokens don't match. - syncing tokens and re-enabling.", zap.String("incomingToken", deviceToken),
			zap.String("endpointToken", *resp.Attributes["Token"]), zap.Bool("EndpointEnabled", endpointEnabled))

		_, err := ns.SetEndpointAttributes(&sns.SetEndpointAttributesInput{
			Attributes: map[string]*string{
				"Token":   aws.String(deviceToken),
				"Enabled": aws.String(strconv.FormatBool(true)),
			},
			EndpointArn: aws.String(platformEndpointARN),
		})

		if err != nil {
			ns.Logger.Error("error setting endpoint attributes", zap.Error(err))
			return "", err
		}
	}

	return platformEndpointARN, nil
}

func (ns *snsAppleNotificationService) createEndpoint(ctx context.Context, deviceToken string) (string, error) {
	endpointRequest := sns.CreatePlatformEndpointInput{
		Token:                  aws.String(deviceToken),
		PlatformApplicationArn: aws.String(ns.platformApplicationARN),
	}
	endpointResponse, err := ns.CreatePlatformEndpoint(&endpointRequest)
	if err != nil {
		return "", err
	}

	return *endpointResponse.EndpointArn, nil
}

func (ns *snsAppleNotificationService) SubscribeTopic(ctx context.Context, impartWealthId, topicARN, platformEndpointARN string) error {
	// user, err := dbmodels.Users(Where("impart_wealth_id = ?", impartWealthId)).One(ctx, ns.db)
	// if err != nil {
	// 	return err
	// }

	currentSubscriptions, err := dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.PlatformEndpointArn.EQ(platformEndpointARN)).One(ctx, ns.db)
	if err != nil {
		return err
	}

	if currentSubscriptions.TopicArn == topicARN {
		//already subbed
		return nil
	}

	subscriptionRequest := sns.SubscribeInput{
		TopicArn:              aws.String(topicARN),
		Endpoint:              aws.String(platformEndpointARN),
		Protocol:              aws.String("application"),
		ReturnSubscriptionArn: aws.Bool(true),
	}

	resp, err := ns.Subscribe(&subscriptionRequest)
	if err != nil {
		ns.Logger.Error("error attempting to subscribe",
			zap.Error(err),
			zap.String("topicARN", topicARN),
			zap.String("platformApplicationARN", platformEndpointARN))
		return err
	}
	p := &dbmodels.NotificationSubscription{
		ImpartWealthID:      impartWealthId,
		TopicArn:            topicARN,
		SubscriptionArn:     *resp.SubscriptionArn,
		PlatformEndpointArn: platformEndpointARN,
	}

	return p.Upsert(ctx, ns.db, boil.Infer(), boil.Infer())
}

func (ns *snsAppleNotificationService) UnsubscribeTopic(ctx context.Context, impartWealthId, SubscriptionARN string) (err error) {
	req := sns.UnsubscribeInput{
		SubscriptionArn: aws.String(SubscriptionARN),
	}

	if _, err = ns.Unsubscribe(&req); err != nil {
		ns.Logger.Error("error attempting to unsubscribe from topic",
			zap.Error(err),
			zap.String("SubscriptionARN", SubscriptionARN))
		//noop, still delete the row from db
	}
	_, err = dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.ImpartWealthID.EQ(impartWealthId),
		dbmodels.NotificationSubscriptionWhere.SubscriptionArn.EQ(SubscriptionARN)).DeleteAll(ctx, ns.db)

	ns.Logger.Error("error attempting to unsubscribe",
		zap.Error(err),
		zap.String("SubscriptionARN", SubscriptionARN))

	return nil
}

func (ns *snsAppleNotificationService) UnsubscribeTopicForDevice(ctx context.Context, impartWealthID, topicARN, platformEndpointARN string) (err error) {
	currentSubscriptions, err := dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.PlatformEndpointArn.EQ(platformEndpointARN)).One(ctx, ns.db)
	if err != nil {
		return err
	}
	req := sns.UnsubscribeInput{
		SubscriptionArn: aws.String(currentSubscriptions.SubscriptionArn),
	}
	if _, err = ns.Unsubscribe(&req); err != nil {
		ns.Logger.Error("error attempting to unsubscribe from topic",
			zap.Error(err),
			zap.String("SubscriptionARN", currentSubscriptions.SubscriptionArn))
		//noop, still delete the row from db
	}
	_, err = dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.PlatformEndpointArn.EQ(platformEndpointARN)).DeleteAll(ctx, ns.db)

	ns.Logger.Error("error attempting to unsubscribe",
		zap.Error(err),
		zap.String("platformEndpointARN", platformEndpointARN))
	return nil
}

func (ns *snsAppleNotificationService) UnsubscribeAll(ctx context.Context, impartWealthID string) error {
	user, err := dbmodels.Users(dbmodels.UserWhere.ImpartWealthID.EQ(impartWealthID)).One(ctx, ns.db)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	currentSubscriptions, err := dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.ImpartWealthID.EQ(user.ImpartWealthID)).All(ctx, ns.db)
	if err != nil {
		return err
	}

	for _, sub := range currentSubscriptions {
		if _, err = ns.Unsubscribe(&sns.UnsubscribeInput{
			SubscriptionArn: aws.String(sub.SubscriptionArn),
		}); err != nil {
			ns.Logger.Error("error attempting to unsubscribe from topic",
				zap.Error(err),
				zap.String("SubscriptionARN", sub.SubscriptionArn))
		}
	}
	_, err = currentSubscriptions.DeleteAll(ctx, ns.db)
	return err
}
