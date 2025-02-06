package s3

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Operator struct {
	// client 是 S3 客戶端。
	Client *s3.Client
	// Bucket 是 S3 存儲桶的名稱。
	Bucket string
	// PublicEndpoint 是 S3 存儲桶的公開 Endpoint。
	PublicEndpoint *url.URL
}

func NewS3Operator(client *s3.Client, bucket, publicBaseURL string) (*S3Operator, error) {
	const op = "NewS3Operator"
	publicEndpoint, err := url.Parse(publicBaseURL)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to parse public base URL, err=%w", op, err)
	}
	return &S3Operator{Client: client, Bucket: bucket, PublicEndpoint: publicEndpoint}, nil
}

func (s *S3Operator) UploadFileToS3(ctx context.Context, path, contentType string, fileContent []byte) (string, error) {
	const op = "UploadFileToS3"
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(path),
		Body:        bytes.NewReader(fileContent),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("[%s] Fail to upload file to S3, err=%w", op, err)
	}
	uri := *s.PublicEndpoint
	uri.Path = path
	return uri.String(), nil
}
