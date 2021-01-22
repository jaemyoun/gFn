package storageFn_test

import (
	"fmt"
	"github.com/jaemyoun/gFn/storageFn"
	"testing"
)

func TestNew(t *testing.T) {

	storageFn.New(storageFn.S3, region)
	storageFn.NewWithBucket(storageFn.GCS, bucket)

	ListOutput := storageFn.List(&storageFn.ListInput{
		bucket:     bucket,
		prefix:     prefix,
		delimiter:  delimiter,
		onlyObject: false,
	})

	select {
	case err := <-ListOutput.Done():
		if err != nil {
			t.Errorf("wrong list objects: %v", err)
		}
	case output := <-ListOutput.Output():
		fmt.Println(output)
	}

}
