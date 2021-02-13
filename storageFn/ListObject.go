package storageFn

import (
	"context"
	"github.com/jaemyoun/gFn/loopFn"
	"time"
)

type ListInput struct {
	Bucket    string
	Prefix    string
	Delimiter string
	OptFn     func(string) []string
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
	Err          error
}

type internalListIO struct {
	input  *ListInput
	output *ListOutput

	storageFn          *StorageFn
	lf                 *loopFn.LoopFn
	undefinedDelimiter bool

	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (c *StorageFn) List(input *ListInput) *ListOutput {
	io := internalListIO{
		input:     input,
		output:    &ListOutput{},
		storageFn: c,
		lf:        loopFn.New(input.Prefix),
	}
	io.ctx, io.ctxCancel = context.WithCancel(context.Background())
	if len(io.input.Delimiter) == 0 {
		io.input.Delimiter = "/"
		io.undefinedDelimiter = true
	}

	switch c.cloud {
	case S3:
		go io.lf.Do(io.s3ListObjects)
	case GCS:
	}

	return io.output
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

func (c ListOutput) Output() chan StorageObject {
	return c.ch
}
