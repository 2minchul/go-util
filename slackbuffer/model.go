package slackbuffer

type slackRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}
