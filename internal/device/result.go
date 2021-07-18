package internal

import (
	"time"
)

type Result struct {
	input     string
	outputs   []string
	err       error
	writeTime time.Duration
	readTime  time.Duration
	totalTime time.Duration
}

func (r Result) Failed() bool {
	return r.err != nil
}

func (r Result) GetError() error {
	return r.err
}

func (r Result) GetOutput() []string {
	return r.outputs
}
