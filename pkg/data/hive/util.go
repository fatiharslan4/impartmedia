package data

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/impartwealthapp/backend/pkg/impart"
)

// DefaultTime is the default value of time.Time
var DefaultTime time.Time

// DefaultLimit is the default number of items that will be returned when getting a list of items
var DefaultLimit int64 = 25

func handleAWSErr(err error) error {
	if err == nil {
		return nil
	}

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeResourceNotFoundException:
			return impart.ErrNotFound
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
		default:
			return err
		}
	}
	return err
}

func conditionalUpdateNoError(err error) error {
	if err == nil {
		return nil
	}

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeConditionalCheckFailedException:
			return nil
		default:
			return err
		}
	}
	return err
}
