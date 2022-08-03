package message

import "encoding/json"

type MessageKind uint32

const (
	MessageKindNone        = MessageKind(iota)
	MessageKindOrderCreate = MessageKind(iota)
	MessageKindOrderCancel = MessageKind(iota)
	MessageKindTrade       = MessageKind(iota)
	MessageKindCancel      = MessageKind(iota)
	NumOfMessageKind       = int(iota)
)

type Message interface {
	GetKind() MessageKind
	GetData() []byte
}

type simpleMessage struct {
	kind MessageKind
	data []byte
}

func NewMessage(kind MessageKind, data interface{}) Message {
	bs, _ := json.Marshal(data)
	return NewMessageWithBytes(kind, bs)
}

func NewMessageWithBytes(kind MessageKind, data []byte) Message {
	return &simpleMessage{
		kind: kind,
		data: data,
	}
}

func (m *simpleMessage) GetKind() MessageKind {
	return m.kind
}

func (m *simpleMessage) GetData() []byte {
	return m.data
}

type AcknowledgementMessage interface {
	Message
	// Ack acknowledges that the message is handled.
	Ack()
	// Nack acknowledges the message is failed to handle.
	Nack()
}

type channelQueueMessage struct {
	Message
	queue *channelQueue
}

func (m *channelQueueMessage) Ack() {
}

func (m *channelQueueMessage) Nack() {
	m.queue.ch <- m
}
