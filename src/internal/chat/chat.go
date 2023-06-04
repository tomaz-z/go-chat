package chat

const (
	ChatAPIName = "GoChat"
)

type Message struct {
	Author  string
	Message string
}

type ChatService interface {
	PostMessage(m Message)
	ReadMessages() <-chan Message
}

func New() ChatService {
	return &service{
		messages: make(chan Message),
	}
}

type service struct {
	messages chan Message
}

func (s *service) PostMessage(m Message) {
	s.messages <- m
}

func (s *service) ReadMessages() <-chan Message {
	return s.messages
}
