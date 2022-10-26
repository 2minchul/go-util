package slackbuffer

import (
	"net/http"
	"time"
)

type Config struct {
	// Interval is the interval to send messages to slack. Default is 1 second.
	Interval time.Duration
	// MessageChannelSize is the size of the message channel. Default is 10.
	MessageChannelSize int
	// Log is the function to log errors. Default is log to stderr.
	// If you want to disable logging, use NoLog.
	Log LogFunc
	// HTTPClient is the http client to send messages to slack. Default is http.DefaultClient.
	HTTPClient *http.Client
	// Timeout is the timeout to send messages to slack. Default is 5 seconds.
	Timeout time.Duration
}

func (c Config) useDefault() Config {
	if c.Interval == 0 {
		c.Interval = time.Second
	}
	if c.MessageChannelSize == 0 {
		c.MessageChannelSize = 10
	}
	if c.Log == nil {
		c.Log = defaultLog
	}
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	return c
}
