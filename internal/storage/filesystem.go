package storage

import (
	"context"
	"fmt"

	"golang.org/x/net/webdav"

	"cloud-webdav-server/internal/config"
)

// New creates a webdav.FileSystem from the given configuration.
func New(cfg *config.Config) (webdav.FileSystem, error) {
	switch cfg.StorageType {
	case config.StorageLocal:
		return NewLocal(cfg.LocalDataPath), nil
	case config.StorageS3:
		return NewS3(context.Background(), S3Config{
			Bucket:    cfg.S3Bucket,
			Region:    cfg.S3Region,
			Prefix:    cfg.S3Prefix,
			Endpoint:  cfg.S3Endpoint,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
		})
	case config.StorageGCS:
		return NewGCS(context.Background(), GCSConfig{
			Bucket:      cfg.GCSBucket,
			Prefix:      cfg.GCSPrefix,
			Credentials: cfg.GCSCredentials,
		})
	case config.StorageAzure:
		return NewAzure(context.Background(), AzureConfig{
			Account:          cfg.AzureAccount,
			Key:              cfg.AzureKey,
			Container:        cfg.AzureContainer,
			Prefix:           cfg.AzurePrefix,
			Endpoint:         cfg.AzureEndpoint,
			ConnectionString: cfg.AzureConnectionString,
		})
	default:
		return nil, fmt.Errorf("storage: unknown type %q", cfg.StorageType)
	}
}
