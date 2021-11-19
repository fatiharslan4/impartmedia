package impart

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"go.uber.org/zap"
)

type EmailService interface {
	EmailSending(ctx context.Context, recipient, template string) error
}
type noopEmailService struct {
}

type sesAppleEmailService struct {
	stage string
	*ses.SES
	*zap.Logger
	db *sql.DB
}

func NewImpartEmailService(db *sql.DB, stage, region string, logger *zap.Logger) EmailService {

	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String(region),
		HTTPClient: NewHttpClient(10 * time.Second),
	})
	if err != nil {
		logger.Fatal("unable to create aws session", zap.Error(err))
	}

	sesAppleEmailService := &sesAppleEmailService{
		stage:  stage,
		Logger: logger,
		SES:    ses.New(sess),
		db:     db,
	}

	logger.Debug("created new EmailService",
		zap.String("stage", stage))

	return sesAppleEmailService
}

func NewNoopEmailService() EmailService {
	return &noopEmailService{}
}

func (ns *noopEmailService) EmailSending(ctx context.Context, recipient, template string) error {
	return nil
}

func (ns *sesAppleEmailService) EmailSending(ctx context.Context, recipient, template string) error {

	sender := "support@impartwealth.com"

	subject := Hive_mail_subject
	textBody := Hive_mail_previewtext
	if template == Waitlist_mail {
		subject = Waitlist_mail_subject
		textBody = Waitlist_mail_previewtext
	}

	// The HTML body for the email.

	htmlBody, err := ioutil.ReadFile(fmt.Sprintf("file://%s", "./schemas/html/"+template+".html"))

	Logger.Info("htmlBody", zap.Any("htmlBody", htmlBody))
	Logger.Info("htmlBody", zap.Any("htmlBody", fmt.Sprintf("file://%s", "./schemas/html/"+template+".html")))

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

func SendAWSEMails(ctx context.Context, db *sql.DB, user *dbmodels.User, mailType string) {
	cfg, _ := config.GetImpart()
	emailSending := NewImpartEmailService(db, string(cfg.Env), cfg.Region, Logger)
	err := emailSending.EmailSending(ctx, user.Email, mailType)
	if err != nil {
		Logger.Error("Hive eamil sending Falied", zap.Any("error", err),
			zap.Any("Email", user.Email))
	}
}
