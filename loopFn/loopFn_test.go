package loopFn

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	trace := ""
	answer := "ABCDE"

	lf := New(1)
	go lf.Do(func(value interface{}) {
		switch value.(int) {
		case 1:
			trace += "A"
			lf.Input(2)
			time.Sleep(time.Second * 1)
			trace += "C"
		case 2:
			trace += "B"
			time.Sleep(time.Second * 2)
			lf.Input(10)
			trace += "D"
		case 10:
			trace += "E"
		}
	})

	for output := range lf.Output() {
		trace += output.(string)
	}

	if trace != answer {
		t.Errorf("wrong ordering: (trace: %v != answer: %v", trace, answer)
	}
}
