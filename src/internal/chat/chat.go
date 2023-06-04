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
	Subscribe() <-chan Message
}

func New() ChatService {
	return &service{
		subscriptions: []chan Message{},
	}
}

type service struct {
	subscriptions []chan Message
}

func (s *service) PostMessage(m Message) {
	for _, s := range s.subscriptions {
		s <- m
	}
}

func (s *service) Subscribe() <-chan Message {
	newSubscription := make(chan Message)
	s.subscriptions = append(s.subscriptions, newSubscription)
	return newSubscription
}
