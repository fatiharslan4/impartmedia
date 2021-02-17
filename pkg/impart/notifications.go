package impart

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"go.uber.org/zap"
)

const notificationProfilePropertyPath = "notificationProfile.awsPlatformEndpointARN"

type NotificationService interface {
	NotifyAppleDevice(data NotificationData, alert Alert, deviceToken, platformEndpointARN string) (sentPlatformEndpointARN string, err error)
	NotifyTopic(data NotificationData, alert Alert, topicARN string) error
	SubscribeTopic(platformApplicationARN, topicARN string) (subscriptionARN string, err error)
	UnsubscribeTopic(subscriptionARN string) (err error)

	// SyncTokenEndpoint is meant to be called when a profiles deviceToken has been updated - this will ensure that the platformApplication
	// has the right device token, and the endpoint is enabled.
	SyncTokenEndpoint(deviceToken, platformEndpointARN string) (string, error)
}

type NotificationData struct {
	EventDatetime time.Time `json:"eventDatetime"`
	PostID        string    `json:"postId,omitempty"`
}

type noopNotificationService struct {
}

func (n noopNotificationService) SyncTokenEndpoint(deviceToken, platformEndpointARN string) (string, error) {
	return "", nil
}

func (n noopNotificationService) NotifyTopic(data NotificationData, alert Alert, topicARN string) error {
	return nil
}

func (n noopNotificationService) SubscribeTopic(platformApplicationARN, topicARN string) (subscriptionARN string, err error) {
	return "", nil
}

func (n noopNotificationService) UnsubscribeTopic(subscriptionARN string) error {
	return nil
}

func (n noopNotificationService) NotifyAppleDevice(data NotificationData, alert Alert, deviceToken, platformEndpointARN string) (string, error) {
	return "", nil
}

func NewNoopNotificationService() NotificationService {
	return &noopNotificationService{}
}

type snsAppleNotificationService struct {
	stage string
	*sns.SNS
	*zap.Logger
	platformApplicationARN string
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

func NewImpartNotificationService(stage, region, platformApplicationARN string, logger *zap.Logger) NotificationService {
	//SNS not available in us-east-2
	if strings.EqualFold(region, "us-east-2") {
		region = "us-east-1"
	}
	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String(region),
		HTTPClient: ImpartHttpClient(10 * time.Second),
	})
	if err != nil {
		logger.Fatal("unable to create aws session", zap.Error(err))
	}

	snsAppleNotificationService := &snsAppleNotificationService{
		stage:                  stage,
		Logger:                 logger,
		SNS:                    sns.New(sess),
		platformApplicationARN: platformApplicationARN,
	}

	logger.Debug("created new NotificationService",
		zap.String("stage", stage),
		zap.String("arn", platformApplicationARN))

	return snsAppleNotificationService
}

func (ns *snsAppleNotificationService) NotifyTopic(data NotificationData, alert Alert, topicARN string) error {
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
	_, err = ns.Publish(input)
	return err
}

func (ns *snsAppleNotificationService) NotifyAppleDevice(data NotificationData, alert Alert, deviceToken, platformEndpointARN string) (string, error) {
	var b []byte
	var err error

	platformEndpointARN, err = ns.SyncTokenEndpoint(deviceToken, platformEndpointARN)
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

func (ns *snsAppleNotificationService) SyncTokenEndpoint(deviceToken, platformEndpointARN string) (string, error) {
	var err error
	// No stored endpoint ARN
	if strings.TrimSpace(platformEndpointARN) == "" {
		ns.Logger.Debug("didn't receive a stored endpoint - attempting to create one.")
		platformEndpointARN, err = ns.createEndpoint(deviceToken)
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
			return ns.createEndpoint(deviceToken)
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

func (ns *snsAppleNotificationService) createEndpoint(deviceToken string) (string, error) {
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

func (ns *snsAppleNotificationService) SubscribeTopic(platformApplicationARN, topicARN string) (subscriptionARN string, err error) {
	subscriptionRequest := sns.SubscribeInput{
		TopicArn:              aws.String(topicARN),
		Endpoint:              aws.String(platformApplicationARN),
		Protocol:              aws.String("application"),
		ReturnSubscriptionArn: aws.Bool(true),
	}

	resp, err := ns.Subscribe(&subscriptionRequest)
	if err != nil {
		ns.Logger.Error("error attempting to subscribe",
			zap.Error(err),
			zap.String("topicARN", topicARN),
			zap.String("platformApplicationARN", platformApplicationARN))
		return subscriptionARN, err
	}
	return *resp.SubscriptionArn, nil
}

func (ns *snsAppleNotificationService) UnsubscribeTopic(SubscriptionARN string) (err error) {
	req := sns.UnsubscribeInput{
		SubscriptionArn: aws.String(SubscriptionARN),
	}

	if _, err = ns.Unsubscribe(&req); err != nil {
		ns.Logger.Error("error attempting to subscribe",
			zap.Error(err),
			zap.String("SubscriptionARN", SubscriptionARN))
		return err
	}
	return nil
}
