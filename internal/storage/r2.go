package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	_ "path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	client *s3.Client
	bucket string
}

var r2 *R2Client

func InitR2() {
	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET")

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		panic("R2 config tidak lengkap di .env")
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint,
			SigningRegion: "auto",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		panic(fmt.Sprintf("Gagal init R2 config: %v", err))
	}

	r2 = &R2Client{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}
}

func GetR2() *R2Client {
	if r2 == nil {
		InitR2()
	}
	return r2
}

// UploadFile mengupload file ke R2 dan return public URL (kalau bucket public) atau key
// UploadFile mengupload file ke R2 dan return KEY (path relatif di bucket)
func (r *R2Client) UploadFile(ctx context.Context, file io.Reader, fileName string) (string, error) {
	key := fmt.Sprintf("bukti/%s/%s-%s", time.Now().Format("2006-01"), time.Now().Format("20060102150405"), fileName)

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	return key, nil  // ← return key saja, bukan URL
}

// GetPresignedURL generate temporary signed URL untuk akses file private
func (r *R2Client) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
    presigner := s3.NewPresignClient(r.client)

    req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(r.bucket),
        Key:    aws.String(key),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = expiry  // ← ini yang benar: Expires (bukan ExpireTime)
    })
    if err != nil {
        return "", fmt.Errorf("gagal generate presigned URL: %w", err)
    }

    return req.URL, nil
}