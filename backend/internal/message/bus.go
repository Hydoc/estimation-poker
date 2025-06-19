package message

type MessageBus interface {
	Dispatch(message Message)
	Subscribe(message string, subscriber chan<- Message)
}

type Bus struct {
	handlers map[string][]chan<- Message
}

func (b *Bus) Dispatch(message Message) {
	handlers := b.handlers[message.Type]
	for _, handler := range handlers {
		handler <- message
	}
}

func (b *Bus) Subscribe(message string, subscriber chan<- Message) {
	b.handlers[message] = append(b.handlers[message], subscriber)
}

func NewBus() MessageBus {
	return &Bus{
		handlers: make(map[string][]chan<- Message),
	}
}
