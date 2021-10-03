package slackwrap

import (
	"github.com/slack-go/slack"
	"net/http"
)

type SlackHandlerFactory struct {
	client             *slack.Client
	signingSecret      *string
	eventHandler       EventHandler
	interactiveHandler InteractiveHandler
	blockActionManager *blockActionManager
}

func NewSlackHandlerFactory(cli *slack.Client, signingSecret *string, eh EventHandler, ih InteractiveHandler) *SlackHandlerFactory {
	return &SlackHandlerFactory{
		client:             cli,
		signingSecret:      signingSecret,
		eventHandler:       eh,
		interactiveHandler: ih,
	}
}

func (s *SlackHandlerFactory) CreateEventEndpoint() http.Handler {
	return &EventEndpoint{
		Client:        s.client,
		Handlers:      []EventHandler{s.blockActionManager, s.eventHandler},
		SigningSecret: s.signingSecret,
	}
}

func (s *SlackHandlerFactory) CreateInteractiveEndpoint() http.Handler {
	return &InteractiveEndpoint{
		Client:   s.client,
		Handlers: []InteractiveHandler{s.blockActionManager, s.interactiveHandler},
	}
}

func (s *SlackHandlerFactory) InitBlockAction(f GetEventIdFromEventFunc) {
	s.blockActionManager = NewBlockActionManager(f)
}

func (s *SlackHandlerFactory) RegisterBlockAction(h BlockActionHandler) {
	s.blockActionManager.Register(h)
}
