package trade

import (
	"context"
	"log"
)

type Recorder interface {
	CreateTradeRecord(ctx context.Context, td Trade) error
}

type stdoutRecorder struct {
}

func NewStdoutRecorder() Recorder {
	return &stdoutRecorder{}
}

func (r *stdoutRecorder) CreateTradeRecord(ctx context.Context, td Trade) error {
	log.Printf("trade: %+v", td)
	return nil
}
