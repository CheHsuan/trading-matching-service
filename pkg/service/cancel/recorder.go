package cancel

import (
	"context"
	"log"
)

type Recorder interface {
	CreateCancelRecord(ctx context.Context, ccl Cancel) error
}

type stdoutRecorder struct {
}

func NewStdoutRecorder() Recorder {
	return &stdoutRecorder{}
}

func (r *stdoutRecorder) CreateCancelRecord(ctx context.Context, ccl Cancel) error {
	log.Printf("cancel: %+v", ccl)
	return nil
}
