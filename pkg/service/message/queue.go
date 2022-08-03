package message

import "context"

// Queue defines the message queue interface.
type Queue interface {
	// Push pushes a message to queue.
	Push(ctx context.Context, msg Message) error
	// Pop pops out a message from queue.
	Pop(ctx context.Context) (AcknowledgementMessage, error)
}

type channelQueue struct {
	ch chan AcknowledgementMessage
}

// NewQueue returns a queue implemented by channel queue.
func NewQueue(queueSize int) Queue {
	return &channelQueue{
		ch: make(chan AcknowledgementMessage, queueSize),
	}
}

func (q *channelQueue) Push(ctx context.Context, msg Message) error {
	ackMsg := &channelQueueMessage{
		Message: msg,
		queue:   q,
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.ch <- ackMsg:
		return nil
	}
}

func (q *channelQueue) Pop(ctx context.Context) (AcknowledgementMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case m := <-q.ch:
		return m, nil
	}
}
