package loopFn

import (
	"context"
	"sync"
	"sync/atomic"
)

type LoopFn struct {
	maxJobCount               int
	waitingAndRunningJobCount int64

	chInput  chan interface{}
	chOutput chan interface{}
	buffer   []interface{}

	metricRunCount int64

	startValue interface{}
}

func New(startValue interface{}) *LoopFn {
	return NewWith(startValue, 100)
}

func NewWith(startValue interface{}, maxJobCount int) *LoopFn {
	return &LoopFn{startValue: startValue, maxJobCount: maxJobCount, buffer: make([]interface{}, 0),
		chInput: make(chan interface{}), chOutput: make(chan interface{})}
}

func (c *LoopFn) Do(handler func(value interface{})) {
	chBuffer := make(chan interface{})
	doneJob := make(chan struct{})
	var wgJobs, wgBufMgt sync.WaitGroup
	ctxBufMgt, ctxBufMgtCancel := context.WithCancel(context.Background())

	// set starting value
	atomic.AddInt64(&c.waitingAndRunningJobCount, 1)
	c.buffer = append(c.buffer, c.startValue)

	wgBufMgt.Add(1)
	go func() { // buffer management
		defer wgBufMgt.Done()

		for ctxBufMgt.Err() == nil {
			if len(c.buffer) == 0 {
				select {
				case in := <-c.chInput:
					c.buffer = append(c.buffer, in)
				case <-ctxBufMgt.Done():
					break
				}
			} else {
				select {
				case in := <-c.chInput:
					c.buffer = append(c.buffer, in)
				case chBuffer <- c.buffer[0]:
					c.buffer = c.buffer[1:]
				case <-ctxBufMgt.Done():
					break
				}
			}
		}
	}()

	wgJobs.Add(c.maxJobCount)
	for i := 0; i < c.maxJobCount; i++ {
		go func() { // run jobs
			defer wgJobs.Done()
			for value := range chBuffer {
				atomic.AddInt64(&c.metricRunCount, 1)
				handler(value)
				doneJob <- struct{}{}
			}
		}()
	}

	go func() {
		for ctxBufMgt.Err() == nil {
			<-doneJob
			if atomic.AddInt64(&c.waitingAndRunningJobCount, -1) == 0 {
				ctxBufMgtCancel() // terminate buffer management
				wgBufMgt.Wait()   // wait until terminating buffer management
			}
		}
		close(c.chInput) // close after terminating buffer management
		close(chBuffer)  // terminate go routines in the pool
		wgJobs.Wait()    // wait until terminating go routines in the pool
		close(c.chOutput)
	}()
}

func (c *LoopFn) Input(v interface{}) {
	atomic.AddInt64(&c.waitingAndRunningJobCount, 1)
	c.chInput <- v
}

func (c *LoopFn) Output() chan interface{} {
	return c.chOutput
}

func (c *LoopFn) GetRunCount() int64 {
	return atomic.LoadInt64(&c.metricRunCount)
}
