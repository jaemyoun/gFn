package storageFn

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
)

func (c *internalListIO) s3ListObjects(value interface{}) {
	prefixes := c.input.doOptionFn(value)

	for _, prefix := range prefixes {
		log.Println("list in", prefix)

		if c.storageFn.s3Service == nil {
			kill(fmt.Errorf("the S3 Service is nil"))
			return
		}
		paginator := s3.NewListObjectsV2Paginator(c.storageFn.s3Service.client, &s3.ListObjectsV2Input{
			Bucket:    aws.String(c.input.Bucket),
			Prefix:    aws.String(prefix),
			Delimiter: aws.String(c.input.Delimiter),
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
						Bucket:   c.input.Bucket,
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
					Bucket:       c.input.Bucket,
				}
			}
		}
	}
}

func (c *internalListIO) Done(err error) {
	c.ctxCancel()
	c.output.ch <- StorageObject{
		Err: err,
	}
	close(c.output.ch)
}
