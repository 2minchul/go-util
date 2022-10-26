package slackbuffer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Slack interface {
	SendMessage(ctx context.Context, message string) error
	AddMessage(message string)
	Close()
}

type slack struct {
	// Channel is slack channel name like `#alert`
	Channel string

	// token is slack token
	token string
	// buffer is a queue of messages that have not yet been sent.
	buffer       []string
	ctx          context.Context
	cancel       context.CancelFunc
	httpClient   *http.Client
	ticker       *time.Ticker
	messageCh    chan string
	log          LogFunc
	doneCh       chan struct{}
	slackTimeout time.Duration
}

// AddMessage append a message to the buffer.
func (s *slack) AddMessage(message string) {
	go func() {
		s.messageCh <- message
	}()
}

// SendMessage sends a message to slack immediately.
func (s *slack) SendMessage(ctx context.Context, message string) error {
	body := slackRequest{
		Channel: s.Channel,
		Text:    message,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal json body: %w", err)
	}
	buff := bytes.NewBuffer(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", buff)
	if err != nil {
		return fmt.Errorf("failed to make http request: %w", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", s.token))
	req.Header.Add("content-type", "application/json; charset=UTF-8")
	if _, err = s.httpClient.Do(req); err != nil {
		return fmt.Errorf("failed to request slack: %w", err)
	}
	return nil
}

// Close stops the background service.
// It must be called when the program exits.
func (s *slack) Close() {
	s.ticker.Stop()
	s.cancel()
	<-s.doneCh
}

// startService is an event loop to run in the background.
func (s *slack) startService() {
	defer close(s.doneCh)
	for {
		select {
		case <-s.ticker.C:
			if err := s.flushMessage(); err != nil {
				s.log(err.Error())
			}
		case message, ok := <-s.messageCh:
			if !ok {
				break
			}
			s.buffer = append(s.buffer, message)
		case <-s.ctx.Done():
			s.flushChan()
			if err := s.flushMessage(); err != nil {
				s.log(err.Error())
			}
			return
		}
	}
}

func (s *slack) flushMessage() error {
	if len(s.buffer) == 0 {
		return nil
	}
	message := strings.Join(s.buffer, "\n")
	s.buffer = s.buffer[:0] // clear buffer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (s *slack) flushChan() {
	for {
		select {
		case message, ok := <-s.messageCh:
			if !ok {
				return
			}
			s.buffer = append(s.buffer, message)
		default:
			return
		}
	}
}

func NewSlackWithConfig(ctx context.Context, token string, channel string, config Config) Slack {
	config = config.useDefault()
	ctx, cancel := context.WithCancel(ctx)
	s := &slack{
		Channel:      channel,
		token:        token,
		buffer:       make([]string, config.MessageChannelSize),
		ctx:          ctx,
		cancel:       cancel,
		httpClient:   config.HTTPClient,
		ticker:       time.NewTicker(config.Interval),
		messageCh:    make(chan string, config.MessageChannelSize),
		log:          config.Log,
		doneCh:       make(chan struct{}),
		slackTimeout: config.Timeout,
	}
	go s.startService()
	return s
}

func NewSlack(ctx context.Context, token string, channel string) Slack {
	return NewSlackWithConfig(ctx, token, channel, Config{})
}
