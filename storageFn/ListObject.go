package storageFn

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jaemyoun/gFn/loopFn"
	"log"
	"time"
)

type ListInput struct {
	Bucket    string
	Prefix    string
	Delimiter string
	OptFn     func(string) []string

	storageFn          *StorageFn
	output             *ListOutput
	lf                 *loopFn.LoopFn
	undefinedDelimiter bool

	ctx       context.Context
	ctxCancel context.CancelFunc
}

type ListOutput struct {
	ch chan StorageObject
}

type StorageObject struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	IsObject     bool
	Bucket       string
}

func (c *StorageFn) List(input *ListInput) *ListOutput {
	input.storageFn = c
	input.output = &ListOutput{}
	input.lf = loopFn.New(input.Prefix)
	input.ctx, input.ctxCancel = context.WithCancel(context.Background())
	if len(input.Delimiter) == 0 {
		input.Delimiter = "/"
		input.undefinedDelimiter = true
	}

	switch c.cloud {
	case S3:
		go input.lf.Do(input.s3ListObjects)
	case GCS:
	}

	for output := range input.lf.Output() {
		fmt.Println(output)
	}
	return input.output
}

func (c *ListInput) s3ListObjects(value interface{}) {
	prefixes := c.doOptionFn(value)

	for _, prefix := range prefixes {
		log.Println("list in", prefix)

		if c.storageFn.s3Service == nil {
			kill(fmt.Errorf("the S3 Service is nil"))
			return
		}
		paginator := s3.NewListObjectsV2Paginator(c.storageFn.s3Service.client, &s3.ListObjectsV2Input{
			Bucket:    aws.String(c.Bucket),
			Prefix:    aws.String(prefix),
			Delimiter: aws.String(c.Delimiter),
		})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(c.ctx)
			if err != nil {
				kill(err)
				return
			}
			for _, dir := range page.CommonPrefixes {
				if c.undefinedDelimiter {
					c.lf.Input(*dir.Prefix)
				} else {
					c.output.ch <- StorageObject{
						IsObject: false,
						Key:      *dir.Prefix,
						Bucket:   c.Bucket,
					}
				}
			}
			for _, obj := range page.Contents {
				c.output.ch <- StorageObject{
					IsObject:     true,
					Key:          *obj.Key,
					Size:         obj.Size,
					ETag:         *obj.ETag,
					LastModified: *obj.LastModified,
					Bucket:       c.Bucket,
				}
			}
		}
	}
}

func (c ListInput) doOptionFn(value interface{}) []string {
	var prefixes []string
	if c.OptFn == nil {
		prefixes = []string{value.(string)}
	} else {
		prefixes = c.OptFn(value.(string))
	}
	return prefixes
}
