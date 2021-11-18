package impart

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
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
	UnsubscribeTopicForAllDevice(ctx context.Context, impartWealthID, topicARN string) (err error)

	// SyncTokenEndpoint is meant to be called when a profiles deviceToken has been updated - this will ensure that the platformApplication
	// has the right device token, and the endpoint is enabled.
	SyncTokenEndpoint(ctx context.Context, deviceToken, platformEndpointARN string) (string, error)
	GetEndPointArn(ctx context.Context, deviceToken, platformEndpointARN string) (string, error)

	CreateNotificationTopic(ctx context.Context, topicARN string) (*sns.CreateTopicOutput, error)
	EmailSending(ctx context.Context, topicARN string) error
}

type NotificationData struct {
	EventDatetime time.Time `json:"eventDatetime"`
	PostID        uint64    `json:"postId,omitempty"`
	CommentID     uint64    `json:"commentId,omitempty"`
	HiveID        uint64    `json:"hiveId,omitempty"`
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

func (n noopNotificationService) GetEndPointArn(ctx context.Context, deviceToken, platformEndpointARN string) (string, error) {
	return "", nil
}

func (n noopNotificationService) UnsubscribeTopicForAllDevice(ctx context.Context, impartWealthID, topicARN string) (err error) {
	return nil
}
func (ns *noopNotificationService) EmailSending(ctx context.Context, topicARN string) error {
	return nil
}

func NewNoopNotificationService() NotificationService {
	return &noopNotificationService{}
}

type snsAppleNotificationService struct {
	stage string
	*sns.SNS
	*ses.SES
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
		SES:                    ses.New(sess),
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
	if err != nil {
		ns.Logger.Error("push-notification : After publish input",
			zap.Any("topicARN", topicARN),
			zap.Error(err),
		)
	}
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
		ns.Logger.Info("push-notification : no active devices found for user",
			zap.Any("msg", alert),
			zap.String("impartWealthID", impartWealthID),
		)
	}

	ns.Logger.Info("push-notification : Sending notifications to", zap.Any("devices", activeDevices))

	// loop through the active devices and send notification
	for _, u := range activeDevices {
		if u.NotifyArn == "" {
			ns.Logger.Error("push-notification : empty device token found for user",
				zap.Any("device", u),
				zap.Any("impartWealthID", impartWealthID),
			)
			continue
		}
		if u.R.UserDevice == nil {
			ns.Logger.Error("push-notification : unable to find user device information",
				zap.Any("device", u),
				zap.Any("impartWealthID", impartWealthID),
			)
			continue
		}
		// user device
		userDevice := u.R.UserDevice
		// var notificationStatus bool

		ns.Logger.Info("push-notification : Initiate notification to",
			zap.Any("device", u),
			zap.Any("impartWealthID", impartWealthID),
		)

		_, err := ns.NotifyAppleDevice(ctx, data, alert, userDevice.DeviceToken, u.NotifyArn)
		if err != nil {
			ns.Logger.Error("push-notification : unable to notify to the device",
				zap.Any("device", userDevice),
				zap.Any("error", err),
			)
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

func (ns *snsAppleNotificationService) GetEndPointArn(ctx context.Context, deviceToken, platformEndpointARN string) (string, error) {
	var err error
	// No stored endpoint ARN
	if strings.TrimSpace(platformEndpointARN) == "" {
		ns.Logger.Debug("didn't receive a stored endpoint - attempting to create one.")
		platformEndpointARN, err = ns.createEndpoint(ctx, deviceToken)
		if err != nil {
			ns.Logger.Error("error creating endpoint", zap.Error(err))
			return "", err
		}

		ns.Logger.Info("platformEndpointARN ",
			zap.Any("deviceToken", deviceToken),
			zap.Any("platformEndpointARN", platformEndpointARN),
		)

	}
	return platformEndpointARN, nil
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

		ns.Logger.Info("platformEndpointARN -didn't receive a stored endpoint",
			zap.Any("deviceToken", deviceToken),
			zap.Any("platformEndpointARN", platformEndpointARN),
		)
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

	currentSubscriptions, err := dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.PlatformEndpointArn.EQ(platformEndpointARN)).All(ctx, ns.db)
	if err != nil {
		return err
	}
	for _, sub := range currentSubscriptions {
		if sub.TopicArn == topicARN {
			//already subbed
			return nil
		}
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

	ns.Logger.Info("SubscribeTopic",
		zap.Any("topicARN", topicARN),
		zap.Any("platformEndpointARN", platformEndpointARN),
		zap.Any("resp", resp),
	)

	p := &dbmodels.NotificationSubscription{
		ImpartWealthID:      impartWealthId,
		TopicArn:            topicARN,
		SubscriptionArn:     *resp.SubscriptionArn,
		PlatformEndpointArn: platformEndpointARN,
	}

	err = p.Upsert(ctx, ns.db, boil.Infer(), boil.Infer())
	return err
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
		dbmodels.NotificationSubscriptionWhere.PlatformEndpointArn.EQ(platformEndpointARN)).All(ctx, ns.db)
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
	_, err = dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.PlatformEndpointArn.EQ(platformEndpointARN)).DeleteAll(ctx, ns.db)
	if err != nil {
		ns.Logger.Error("error attempting to unsubscribe",
			zap.Error(err),
			zap.String("platformEndpointARN", platformEndpointARN))
	}
	return nil
}

func (ns *snsAppleNotificationService) UnsubscribeTopicForAllDevice(ctx context.Context, impartWealthID, topicARN string) (err error) {
	currentSubscriptions, err := dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.ImpartWealthID.EQ(impartWealthID)).All(ctx, ns.db)
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
	_, err = dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.ImpartWealthID.EQ(impartWealthID)).DeleteAll(ctx, ns.db)
	if err != nil {
		ns.Logger.Error("error attempting to unsubscribe",
			zap.Error(err),
			zap.String("Impart wealth Id", impartWealthID))

	}
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

func (n noopNotificationService) CreateNotificationTopic(ctx context.Context, topicARN string) (*sns.CreateTopicOutput, error) {
	return nil, nil
}

/// Create topic for each hive
func (ns *snsAppleNotificationService) CreateNotificationTopic(ctx context.Context, topicARN string) (*sns.CreateTopicOutput, error) {
	topic := &topicARN
	input := &sns.CreateTopicInput{Name: topic}
	topicOutput, err := ns.CreateTopic(input)
	if err != nil {
		ns.Logger.Error("error attempting to create topic",
			zap.Error(err),
			zap.String("topicInput", topicARN))

		return nil, err
	}
	return topicOutput, nil
}

const title = "This Week’s Activity"
const titleMostPopularPost = "This Week’s Trending Post"
const bodyMostPopularPost = "Check out the most popular post in your Hive this week"

func NotifyWeeklyActivity(db *sql.DB, logger *zap.Logger) {
	lastweekTime := CurrentUTC().AddDate(0, 0, -7)
	type PostCount struct {
		HiveID               uint64      `json:"hive_id"`
		Post                 uint64      `json:"post"`
		NotificationTopicArn null.String `json:"notification_topic_arn"`
	}
	var weeklyPosts []PostCount
	err := queries.Raw(`
		select count(post_id) as post , post.hive_id as hive_id , hive.notification_topic_arn
		from post
		join hive on post.hive_id=hive.hive_id and hive.deleted_at is null
		where post.deleted_at is null
		and hive.deleted_at is null
		and post.created_at between ? and ?
		group by hive_id
		having count(post_id)>=3 ;
	 `, lastweekTime, CurrentUTC()).Bind(context.TODO(), db, &weeklyPosts)

	if err != nil {
		logger.Error("error while fetching data ", zap.Error(err))
		return
	}
	logger.Info("NotifyWeeklyActivity fetching completed", zap.Any("data", weeklyPosts))
	cfg, _ := config.GetImpart()
	notification := NewImpartNotificationService(db, string(cfg.Env), cfg.Region, cfg.IOSNotificationARN, logger)
	for _, hive := range weeklyPosts {
		pushNotification := Alert{
			Title: aws.String(title),
			Body: aws.String(
				fmt.Sprintf("Check out %d new posts in your Hive this week", hive.Post),
			),
		}
		additionalData := NotificationData{
			EventDatetime: CurrentUTC(),
			HiveID:        hive.HiveID,
		}
		logger.Info("Notification", zap.Any("hive", hive))
		Logger.Info("Notification",
			zap.Any("pushNotification", pushNotification),
			zap.Any("additionalData", additionalData),
			zap.Any("hive", hive),
		)
		err = notification.NotifyTopic(context.TODO(), additionalData, pushNotification, hive.NotificationTopicArn.String)
		if err != nil {
			logger.Error("error sending notification to topic", zap.Error(err))
		}
	}

}

func NotifyWeeklyMostPopularPost(db *sql.DB, logger *zap.Logger) {
	logger.Info("NotifyWeeklyMostPopularPost- start")
	lastweekTime := CurrentUTC().AddDate(0, 0, -7)

	type PostCount struct {
		PostID               uint64      `json:"post_id"`
		TotalActivity        uint64      `json:"totalActivity"`
		NotificationTopicArn null.String `json:"notification_topic_arn"`
	}
	var popularPosts []PostCount
	err := queries.Raw(`
	select * from ( select post_id, post.hive_id as hive_id , hive.notification_topic_arn,
		post.up_vote_count,
		post.comment_count,
		post.up_vote_count+post.comment_count as totalActivity
		from post
		join hive on post.hive_id=hive.hive_id and hive.deleted_at is null
		where post.deleted_at is null
		and hive.deleted_at is null
		and (post.up_vote_count+post.comment_count)>0
		and post.created_at between ? and ?
		group by hive_id,post_id
		order by  totalActivity desc ,post_id desc) as postdata
		group by hive_id
		; 
	`, lastweekTime, CurrentUTC()).Bind(context.TODO(), db, &popularPosts)

	if err != nil {
		logger.Error("error while fetching data ", zap.Error(err))
		return
	}
	logger.Info("Data fetching completed ", zap.Any("popularPosts", popularPosts))
	cfg, _ := config.GetImpart()
	if cfg.Env != config.Local {
		notification := NewImpartNotificationService(db, string(cfg.Env), cfg.Region, cfg.IOSNotificationARN, logger)
		logger.Info("NotifyWeeklyMostPopularPost- fetching complted")
		for _, hive := range popularPosts {
			logger.Info("NotifyWeeklyMostPopularPost-Post Details", zap.Any("hive", hive))
			pushNotification := Alert{
				Title: aws.String(titleMostPopularPost),
				Body:  aws.String(bodyMostPopularPost),
			}
			additionalData := NotificationData{
				EventDatetime: CurrentUTC(),
				PostID:        hive.PostID,
			}
			Logger.Info("NotifyWeeklyMostPopularPost",
				zap.Any("pushNotification", pushNotification),
				zap.Any("additionalData", additionalData),
				zap.Any("hive", hive),
			)
			err = notification.NotifyTopic(context.TODO(), additionalData, pushNotification, hive.NotificationTopicArn.String)
			if err != nil {
				logger.Error("error sending notification to topic", zap.Error(err))
			}

		}
	}
}

func (ns *snsAppleNotificationService) EmailSending(ctx context.Context, topicARN string) error {

	fmt.Println("start")
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	// sender := "success@simulator.amazonses.com"
	sender := "monika.prakash@naicoits.com"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	recipient := "monika.prakash@naicoits.com"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	subject := "Amazon SES Test (AWS SDK for Go)"

	// The HTML body for the email.

	// htmlBody, err := ioutil.ReadFile(fmt.Sprintf("file://%s", "./schemas/html/hive_email.html"))

	htmlBody, err := ioutil.ReadFile(fmt.Sprintf("D:/ImpartWealthApp/back-End/schemas/html/hive_email.html"))

	// htmlBody := "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
	// 	"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
	// 	"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>.</p>"

	//The email body for recipients with non-HTML email clients.
	textBody := "This email was sent with Amazon SES using the AWS SDK for Go."

	// The character encoding for the email.
	charSet := "UTF-8"

	// Assemble the email.

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(string(htmlBody)),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := ns.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		fmt.Println("erororroror")
		fmt.Println(err)
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				Logger.Error("Error", zap.Any("aerr", err))
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
				return err
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				Logger.Error("Error", zap.Any("aerr", err))
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
				Logger.Error("Error", zap.Any("aerr", err))
				return err
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				Logger.Error("Error", zap.Any("aerr", err))
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
				return err
			default:
				Logger.Error("Error", zap.Any("aerr", err))
				fmt.Println(aerr.Error())
				return err
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			Logger.Error("Error", zap.Any("aerr", err))
			fmt.Println(err.Error())
			return err
		}
	}

	fmt.Println("Email Sent to address: " + recipient)
	fmt.Println(result)
	return nil
}
