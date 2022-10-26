package slackbuffer

import (
	"fmt"
	"os"
)

type LogFunc func(message string)

func defaultLog(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
}

func NoLog(message string) {
}
