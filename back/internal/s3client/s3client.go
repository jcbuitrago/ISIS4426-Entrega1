package s3client

import (
	"context"
	"io"
	"time"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client          *s3.Client
	uploadsBucket   string
	processedBucket string
}

func New(uploadsBucket, processedBucket string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return &S3Client{
		client:          s3.NewFromConfig(cfg),
		uploadsBucket:   uploadsBucket,
		processedBucket: processedBucket,
	}, nil
}

func (s *S3Client) DownloadFile(ctx context.Context, key string, bucket string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (s *S3Client) UploadFile(ctx context.Context, key string, bucket string, body io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

// UploadToUploads uploads a file to the uploads bucket
func (s *S3Client) UploadToUploads(ctx context.Context, key string, body io.Reader) error {
	return s.UploadFile(ctx, key, s.uploadsBucket, body)
}

// UploadToProcessed uploads a file to the processed bucket
func (s *S3Client) UploadToProcessed(ctx context.Context, key string, body io.Reader) error {
	return s.UploadFile(ctx, key, s.processedBucket, body)
}

// DownloadFromUploads downloads a file from the uploads bucket
func (s *S3Client) DownloadFromUploads(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.DownloadFile(ctx, key, s.uploadsBucket)
}

// DownloadFromProcessed downloads a file from the processed bucket
func (s *S3Client) DownloadFromProcessed(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.DownloadFile(ctx, key, s.processedBucket)
}

// GeneratePresignedURL generates a presigned URL for downloading a file
func (s *S3Client) GeneratePresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", err
	}
	
	return request.URL, nil
}

// GenerateUploadPresignedURL generates a presigned URL for uploading a file
func (s *S3Client) GenerateUploadPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", err
	}
	
	return request.URL, nil
}

// DeleteFile deletes a file from S3
func (s *S3Client) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// FileExists checks if a file exists in S3
func (s *S3Client) FileExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}
	return true, nil
}

// GetUploadsBucket returns the uploads bucket name
func (s *S3Client) GetUploadsBucket() string {
	return s.uploadsBucket
}

// GetProcessedBucket returns the processed bucket name
func (s *S3Client) GetProcessedBucket() string {
	return s.processedBucket
}
