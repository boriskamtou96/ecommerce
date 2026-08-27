package providers

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "ecommerce/internal/config"
)

// objectCacheControl is safe because every key is content addressed:
// a new upload always produces a new key, so a cached object never
// becomes stale.
const objectCacheControl = "public, max-age=31536000, immutable"

type S3Provider struct {
	client   *s3.Client
	uploader *manager.Uploader
	bucket   string
	endpoint string
}

func NewS3Provider(cfg *appconfig.Config) (*S3Provider, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWS.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.AccessKeyID,
			cfg.AWS.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	// A custom endpoint (LocalStack, MinIO) also implies path style
	// addressing. Real AWS keeps the default virtual hosted style.
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		if cfg.AWS.S3Endpoint != "" {
			options.BaseEndpoint = aws.String(cfg.AWS.S3Endpoint)
			options.UsePathStyle = true
		}
	})

	return &S3Provider{
		client:   client,
		uploader: manager.NewUploader(client),
		bucket:   cfg.AWS.S3Bucket,
		endpoint: cfg.AWS.S3Endpoint,
	}, nil
}

func (p *S3Provider) UploadFile(
	ctx context.Context,
	file *multipart.FileHeader,
	key, contentType string,
) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := p.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(p.bucket),
		Key:          aws.String(key),
		Body:         src,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(objectCacheControl),
	})
	if err != nil {
		return "", err
	}

	return *result.Key, nil
}

func (p *S3Provider) DeleteFile(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	})
	return err
}
