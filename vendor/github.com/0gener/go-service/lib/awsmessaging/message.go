package awsmessaging

type MessageOpt func(*Message)
type MessageAttributes map[string]string

type Message struct {
	Data       []byte
	Attributes MessageAttributes
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
