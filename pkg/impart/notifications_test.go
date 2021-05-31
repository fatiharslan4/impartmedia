package impart

var nd NotificationData = NotificationData{
	EventDatetime: CurrentUTC(),
}

//TODO: Uncomment once we fix SNS expired certs
// func TestImpartNotificationServiceDev(t *testing.T) {
// 	cfg, err := config.GetImpart()
// 	if err != nil {
// 		panic(err)
// 	}
// 	db, err := cfg.GetDBConnection()

// 	if err != nil {
// 		panic(err)
// 	}
// 	logger, _ := zap.NewDevelopment()
// 	svc := NewImpartNotificationService(db, "dev", "us-east-2", "arn:aws:sns:us-east-1:518740895671:SNSHiveNotification", logger)

// 	_, err = svc.NotifyAppleDevice(context.TODO(), nd, Alert{Title: aws.String("unit test title"), Body: aws.String("Unit Test Body")}, "77725527-06ff-4f20-8428-670ae24362fd", "")
// 	assert.NoError(t, err)
// }

//
//func TestImpartNotificationServiceProd(t *testing.T) {
//	logger, _ := zap.NewDevelopment()
//	svc := NewImpartNotificationService("preprod", "us-east-2", "arn:aws:sns:us-east-1:518740895671:app/APNS/impart_wealth", logger)
//
//	_, err := svc.NotifyAppleDevice(nd, Alert{Title: aws.String("Unit Test Title"), Body: aws.String("Unit Test Body")}, "6dc0b5dfb805755c4e3d13d09cd4c509f680454691df55251b9ca7f231e189fa", "")
//	assert.NoError(t, err)
//}
//
//func TestImpartNotificationServiceProdWithEndpoint(t *testing.T) {
//	logger, _ := zap.NewDevelopment()
//	svc := NewImpartNotificationService("preprod", "us-east-2", "arn:aws:sns:us-east-1:518740895671:app/APNS/impart_wealth", logger)
//
//	_, err := svc.NotifyAppleDevice(nd, Alert{Title: aws.String("unit test title"), Body: aws.String("Unit Test Body")}, "6dc0b5dfb805755c4e3d13d09cd4c509f680454691df55251b9ca7f231e189fa",
//		"arn:aws:sns:us-east-1:518740895671:endpoint/APNS/impart_wealth/23ab0556-1c5d-3252-85cd-9cab32ea62c6")
//	assert.NoError(t, err)
//}

// func TestNotifyTopic(t *testing.T) {
// 	cfg, err := config.GetImpart()
// 	if err != nil {
// 		panic(err)
// 	}
// 	db, err := cfg.GetDBConnection()

// 	if err != nil {
// 		panic(err)
// 	}
// 	logger, _ := zap.NewDevelopment()
// 	svc := NewImpartNotificationService(db, "dev", "us-east-2", "arn:aws:sns:us-east-1:340593047560:TestTopic", logger)

// 	pushNotification := Alert{
// 		Title: aws.String("Test Title"),
// 		Body:  aws.String("Test subject"),
// 	}

// 	additionalData := NotificationData{
// 		EventDatetime: CurrentUTC(),
// 		PostID:        1,
// 	}

// 	err = svc.NotifyTopic(context.TODO(), additionalData, pushNotification, "arn:aws:sns:us-east-1:340593047560:TestTopic:5aca0eb4-accc-42dc-a899-edc20e082654")
// 	assert.NoError(t, err)
// }
