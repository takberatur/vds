package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
)

type StorageClient interface {
	UploadFile(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error)
	GetFileURL(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (string, error)
	DeleteFile(ctx context.Context, bucketName string, objectName string) error
	CreateBucket(ctx context.Context, bucketName string) error
}

type minioClient struct {
	client   *minio.Client
	endpoint string
	useSSL   bool
}

func NewStorageClient(endpoint, accessKey, secretKey string, useSSL bool) (StorageClient, error) {
	parsedURL, err := url.Parse(endpoint)
	if err == nil && parsedURL.Host != "" {
		endpoint = parsedURL.Host
	}

	// Remove scheme prefix if present in endpoint variable (just in case url.Parse failed or left it)
	// MinIO client expects "hostname:port" or "hostname" without scheme
	if len(endpoint) > 0 {
		// remove https:// or http:// prefix
		if len(endpoint) >= 8 && endpoint[:8] == "https://" {
			endpoint = endpoint[8:]
		} else if len(endpoint) >= 7 && endpoint[:7] == "http://" {
			endpoint = endpoint[7:]
		}
	}

	minioClientObj, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       useSSL,
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &minioClient{
		client:   minioClientObj,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

func (c *minioClient) CreateBucket(ctx context.Context, bucketName string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	exists, err := c.client.BucketExists(subCtx, bucketName)
	if err != nil {
		if errors.Is(subCtx.Err(), context.DeadlineExceeded) {
			log.Error().Err(err).Str("bucket", bucketName).Msg("MinIO CreateBucket BucketExists timed out")
		}
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = c.client.MakeBucket(subCtx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			if errors.Is(subCtx.Err(), context.DeadlineExceeded) {
				log.Error().Err(err).Str("bucket", bucketName).Msg("MinIO CreateBucket MakeBucket timed out")
			}
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", bucketName).Msg("Bucket created successfully")

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"AWS": ["*"]
					},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

		if err := c.client.SetBucketPolicy(subCtx, bucketName, policy); err != nil {
			if errors.Is(subCtx.Err(), context.DeadlineExceeded) {
				log.Error().Err(err).Str("bucket", bucketName).Msg("MinIO CreateBucket SetBucketPolicy timed out")
			}
			log.Error().Err(err).Str("bucket", bucketName).Msg("Failed to set bucket policy")
		} else {
			log.Info().Str("bucket", bucketName).Msg("Bucket policy set to public read")
		}
	} else {
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"AWS": ["*"]
					},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

		if err := c.client.SetBucketPolicy(subCtx, bucketName, policy); err != nil {
			if errors.Is(subCtx.Err(), context.DeadlineExceeded) {
				log.Error().Err(err).Str("bucket", bucketName).Msg("MinIO CreateBucket ensure policy timed out")
			}
			log.Error().Err(err).Str("bucket", bucketName).Msg("Failed to ensure bucket policy")
		}
	}
	return nil
}

func (c *minioClient) UploadFile(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	info, err := c.client.PutObject(subCtx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	log.Debug().Str("bucket", bucketName).Str("object", objectName).Int64("size", info.Size).Msg("Successfully uploaded file")

	protocol := "http"
	if c.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, c.endpoint, bucketName, objectName), nil
}

func (c *minioClient) GetFileURL(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (string, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	reqParams := make(url.Values)
	presignedURL, err := c.client.PresignedGetObject(subCtx, bucketName, objectName, expiry, reqParams)
	if err != nil {
		if errors.Is(subCtx.Err(), context.DeadlineExceeded) {
			log.Error().Err(err).Str("bucket", bucketName).Str("object", objectName).Msg("MinIO GetFileURL timed out")
		}
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}
	return presignedURL.String(), nil
}

func (c *minioClient) DeleteFile(ctx context.Context, bucketName string, objectName string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	err := c.client.RemoveObject(subCtx, bucketName, objectName, opts)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
