package engine

import (
	"context"
	"encoding/json"
	cancelsvc "trading-matching-service/pkg/service/cancel"
	msgsvc "trading-matching-service/pkg/service/message"
)

type cancelEngine struct {
	cancelQ        msgsvc.Queue
	cancelRecorder cancelsvc.Recorder
}

// NewCancelEngined return a cancel engine.
func NewCancelEngine(cancelQ msgsvc.Queue, cancelRecorder cancelsvc.Recorder) Engine {
	return &cancelEngine{
		cancelQ:        cancelQ,
		cancelRecorder: cancelRecorder,
	}
}

func (e *cancelEngine) Run(ctx context.Context) error {
	for {
		msg, err := e.cancelQ.Pop(ctx)
		if err != nil {
			return err
		}
		e.handle(ctx, msg)
	}
}

func (e *cancelEngine) handle(ctx context.Context, msg msgsvc.AcknowledgementMessage) {
	if msg.GetKind() != msgsvc.MessageKindCancel {
		return
	}

	bs := msg.GetData()
	ccl := cancelsvc.Cancel{}
	if err := json.Unmarshal(bs, &ccl); err != nil {
		// not a valid message, drop it
		msg.Ack()
		return
	}

	if err := e.cancelRecorder.CreateCancelRecord(ctx, ccl); err != nil {
		msg.Nack()
		return
	}

	msg.Ack()
}
