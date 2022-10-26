# Slack Buffer

## Example

```go
package main

import (
	"context"
	"time"

	"github.com/2minchul/go-util/slackbuffer"
)

func main() {
	ctx := context.Background()
	s := slackbuffer.NewSlack(ctx, "xoxb-**************************", "#my-channel")
	defer s.Close()
	s.AddMessage("hello")
	s.AddMessage("world")
	time.Sleep(2 * time.Second)
	s.AddMessage("hi")
	time.Sleep(2 * time.Second)
	s.AddMessage("bye")
}
```
