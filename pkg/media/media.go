package media

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/models"
)

//
type S3Storage struct {
	BucketName   string
	BucketRegion string
}
type StorageConfigurations struct {
	Storage   string // with local / s3
	MediaPath string
	S3Storage
}

//uploader
type FileUpload struct {
	StorageConfigurations
}

func New(opt StorageConfigurations) *FileUpload {
	return &FileUpload{
		StorageConfigurations: opt,
	}
}

func LoadMediaConfig(cfg *config.Impart) StorageConfigurations {
	sc := StorageConfigurations{}

	if v, ok := cfg.Media["Storage"]; ok {
		sc.Storage = v
	}
	if v, ok := cfg.Media["BasePath"]; ok {
		sc.MediaPath = v
	}
	if v, ok := cfg.Media["Bucket"]; ok {
		sc.BucketName = v
	}

	sc.BucketRegion = cfg.Region
	return sc
}

type s3Uploader struct {
	StorageConfigurations
}

// media service
type MediaUploadService interface {
	UploadMultipleFile(files []models.File) ([]models.File, error)
}

// Upload multiple files
func (fp *FileUpload) UploadMultipleFile(files []models.File) ([]models.File, error) {
	var uploaderService MediaUploadService

	switch fp.Storage {
	case "s3":
		uploaderService = &s3Uploader{
			StorageConfigurations: fp.StorageConfigurations,
		}
	default:
		return []models.File{}, fmt.Errorf("unable to identify media storage")
	}

	return uploaderService.UploadMultipleFile(files)
}

// Upload multiple files
func (up *s3Uploader) UploadMultipleFile(files []models.File) ([]models.File, error) {
	var uploadedFiles []models.File
	// errorchan := make(chan error)
	uploadedFilesChan := make(chan models.File)

	s, err := up.NewSession()
	if err != nil {
		return []models.File{}, err
	}

	var wg sync.WaitGroup
	wg.Add(len(files))

	// upload the files
	for _, f := range files {
		fileObj := f

		go func() {
			defer wg.Done()
			file, _ := up.UploadSingleFile(s, fileObj)
			uploadedFilesChan <- file
		}()
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(uploadedFilesChan)
	}()

	for {
		res, ok := <-uploadedFilesChan
		if !ok {
			fmt.Println("Channel Close ", ok)
			break
		}
		uploadedFiles = append(uploadedFiles, res)
	}

	return uploadedFiles, nil
}

//
// Upload single file to s3
//
func (up *s3Uploader) UploadSingleFile(s *session.Session, file models.File) (models.File, error) {
	fileName := file.FileName
	ipFile, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return models.File{}, err
	}

	srcFile := bytes.NewReader(ipFile)
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(up.BucketName),
		Key:    aws.String(fileName),
		Body:   srcFile,
	})
	if err != nil {
		return models.File{}, err
	}

	//append the url to the response
	file.URL = up.ConstructFilePath(fileName)
	return file, nil
}

// create new session for s3
func (up *s3Uploader) NewSession() (*session.Session, error) {
	s, err := session.NewSession(&aws.Config{Region: aws.String(up.BucketRegion)})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// construct uploded file path
func (up *s3Uploader) ConstructFilePath(filename string) string {
	return fmt.Sprintf(" https://%s.s3.%s.amazonaws.com/%s", up.BucketName, up.BucketRegion, filename)
}
