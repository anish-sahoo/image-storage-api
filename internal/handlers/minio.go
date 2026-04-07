package handlers

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client
var S3Bucket string

// InitS3 connects to an S3-compatible store (MinIO locally, AWS S3 in prod).
// To switch to real AWS S3: set endpoint="" and usePathStyle=false, remove
// static credentials, and rely on the standard AWS credential chain instead.
func InitS3(endpoint, accessKey, secretKey, bucket string, useSSL bool) error {
	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	cfg := aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	}

	S3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(scheme + "://" + endpoint)
			o.UsePathStyle = true // required for MinIO
		}
	})
	S3Bucket = bucket

	// Ensure bucket exists
	ctx := context.Background()
	_, err := S3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		_, err = S3Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
		return err
	}
	return nil
}

func s3Upload(objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := S3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(S3Bucket),
		Key:           aws.String(objectKey),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	return err
}

func s3GetObject(objectKey string) (io.ReadCloser, error) {
	resp, err := S3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(S3Bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func s3DeleteObject(objectKey string) error {
	_, err := S3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(S3Bucket),
		Key:    aws.String(objectKey),
	})
	return err
}
