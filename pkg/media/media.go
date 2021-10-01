package media

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
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

// s3 uploader
type s3Uploader struct {
	StorageConfigurations
}

// s3 uploader
type localUploader struct {
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
	case "", "local":
		uploaderService = &localUploader{
			StorageConfigurations: fp.StorageConfigurations,
		}
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
	errChan := make(chan error)
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
			file, err := up.UploadSingleFile(s, fileObj)
			if err != nil {
				errChan <- err
			}
			uploadedFilesChan <- file
		}()
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(uploadedFilesChan)
		close(errChan)
	}()

	select {
	case file := <-uploadedFilesChan:
		uploadedFiles = append(uploadedFiles, file)
	case err := <-errChan:
		return uploadedFiles, err
	}
	return uploadedFiles, nil
}

//
// Upload single file to s3
//
func (up *s3Uploader) UploadSingleFile(s *session.Session, file models.File) (models.File, error) {

	fileName := file.FileName
	// append base path
	fileName = fmt.Sprintf("%s%s%s", up.MediaPath, file.FilePath, fileName)

	ipFile, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return models.File{}, err
	}

	srcFile := bytes.NewReader(ipFile)
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(up.BucketName),
		Key:         aws.String(fileName),
		Body:        srcFile,
		ContentType: &file.FileType,
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return models.File{}, err
	}
	//append the url to the response
	file.URL = up.ConstructS3FilePath(fileName)
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
func (up *s3Uploader) ConstructS3FilePath(filename string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com%s", up.BucketName, up.BucketRegion, filename)
}

// local
func (lp *localUploader) UploadMultipleFile(files []models.File) ([]models.File, error) {
	var uploadedFiles []models.File
	errChan := make(chan error)
	uploadedFilesChan := make(chan models.File)

	var wg sync.WaitGroup
	wg.Add(len(files))

	// upload the files
	for _, f := range files {
		fileObj := f

		go func() {
			defer wg.Done()
			file, err := lp.UploadSingleFile(fileObj)
			if err != nil {
				errChan <- err
			}
			uploadedFilesChan <- file
		}()
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(uploadedFilesChan)
		close(errChan)
	}()

	select {
	case file := <-uploadedFilesChan:
		uploadedFiles = append(uploadedFiles, file)
	case err := <-errChan:
		return uploadedFiles, err
	}
	return uploadedFiles, nil
}

//
// Upload single file to local
//
func (lp *localUploader) UploadSingleFile(file models.File) (models.File, error) {
	fileName := file.FileName

	// append base path
	fileName = fmt.Sprintf("%s%s%s", lp.MediaPath, file.FilePath, fileName)

	ipFile, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return models.File{}, err
	}

	imgFile, err := os.Create(fileName)
	if err != nil {
		return models.File{}, err
	}
	srcFile := bytes.NewReader(ipFile)
	_, err = io.Copy(imgFile, srcFile)

	if err != nil {
		return models.File{}, err
	}

	//append the url to the response
	file.URL = lp.ConstructLocalFilePath(fileName)
	return file, nil
}

// construct uploded file path
func (lp *localUploader) ConstructLocalFilePath(filename string) string {
	return filename
}
