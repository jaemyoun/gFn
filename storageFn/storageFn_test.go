package storageFn_test

import (
	"context"
	"fmt"
	"github.com/jaemyoun/gFn/storageFn"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {

	region := "ap-northeast-1"
	s, err := storageFn.New(storageFn.S3, region)
	if err != nil {
		t.Error(err)
	}
	bucket := "dev-"
	storageFn.NewWithBucket(storageFn.GCS, bucket)

	out := s.List(&storageFn.ListInput{
		Bucket:    bucket,
		Prefix:    "server/",
		Delimiter: "/",
	})

	select {
	case err := <-out.Done():
		if err != nil {
			t.Errorf("wrong list objects: %v", err)
		}
	case output := <-out.Output():
		fmt.Println(output)
	}

}
