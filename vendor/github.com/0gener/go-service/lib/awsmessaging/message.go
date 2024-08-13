package awsmessaging

import "github.com/google/uuid"

type MessageOpt func(*Message)
type MessageAttributes map[string]string

type Message struct {
	ID         uuid.UUID
	Data       []byte
	Attributes MessageAttributes
	Err        error
}

func WithAttribute(key string, val string) MessageOpt {
	return func(message *Message) {
		message.Attributes[key] = val
	}
}

func NewMessage(data []byte, opts ...MessageOpt) *Message {
	msg := &Message{
		Data:       data,
		Attributes: make(MessageAttributes),
	}

	for _, opt := range opts {
		opt(msg)
	}

	return msg
}
