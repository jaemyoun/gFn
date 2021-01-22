package storageFn

import (
	"context"
	"time"
)

type StorageFn struct {
	cloud     kindOfCloud
	s3Service *s3Service
}

type kindOfCloud int

const (
	S3 kindOfCloud = iota
	GCS
)

func New(cloud kindOfCloud, region string) (*StorageFn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	switch cloud {
	case S3:
		svc, err := newS3(ctx, region)
		if err != nil {
			return nil, err
		}
		return &StorageFn{cloud: cloud, s3Service: svc}, nil
	case GCS:
	}
	return nil, nil
}

func NewWithBucket(cloud kindOfCloud, bucket string) (s *StorageFn, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	var region string
	switch cloud {
	case S3:
		region, err = getS3RegionFrom(ctx, bucket)
		if err != nil {
			return nil, err
		}
	case GCS:
	}
	return New(cloud, region)
}
