package engine

import (
	"context"
	"encoding/json"
	msgsvc "trading-matching-service/pkg/service/message"
	tradesvc "trading-matching-service/pkg/service/trade"
)

type tradeEngine struct {
	tradeQ        msgsvc.Queue
	tradeRecorder tradesvc.Recorder
}

// NewTradeEngined return a trade engine.
func NewTradeEngine(tradeQ msgsvc.Queue, tradeRecorder tradesvc.Recorder) Engine {
	return &tradeEngine{
		tradeQ:        tradeQ,
		tradeRecorder: tradeRecorder,
	}
}

func (e *tradeEngine) Run(ctx context.Context) error {
	for {
		msg, err := e.tradeQ.Pop(ctx)
		if err != nil {
			return err
		}
		e.handle(ctx, msg)
	}
}

func (e *tradeEngine) handle(ctx context.Context, msg msgsvc.AcknowledgementMessage) {
	if msg.GetKind() != msgsvc.MessageKindTrade {
		return
	}

	bs := msg.GetData()
	td := tradesvc.Trade{}
	if err := json.Unmarshal(bs, &td); err != nil {
		// not a valid message, drop it
		msg.Ack()
		return
	}

	if err := e.tradeRecorder.CreateTradeRecord(ctx, td); err != nil {
		msg.Nack()
		return
	}

	msg.Ack()
}
