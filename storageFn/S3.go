package storageFn

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Service struct {
	client *s3.Client
}

func newS3(ctx context.Context, region string) (*s3Service, error) {
	cfg, err := getS3Config(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Region = region
	return &s3Service{client: s3.NewFromConfig(cfg)}, nil
}

func getS3RegionFrom(ctx context.Context, bucket string) (string, error) {
	cfg, err := getS3Config(ctx)
	if err != nil {
		return "", err
	}
	region, err := manager.GetBucketRegion(ctx, s3.NewFromConfig(cfg), bucket)
	if err != nil {
		var bnf manager.BucketNotFound
		if errors.As(err, &bnf) {
			return "", fmt.Errorf("unable to find bucket %s's region", bucket)
		}
		return "", fmt.Errorf("failed to set S3 region: %v", err)
	}
	return region, nil
}

func getS3Config(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load config, %v", err)
	}
	return cfg, nil
}
