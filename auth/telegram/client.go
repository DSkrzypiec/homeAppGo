package telegram

import (
	"net/http"
	"strconv"
)

// Client is a Telegram client which handles communication with Telegram
// channel.
type Client struct {
	httpClient *http.Client
	botToken   string
	channelId  int64
}

// NewClient instantiates new client.
func NewClient(httpClient *http.Client, botToken string, channelId string) *Client {
	chatIdInt, _ := strconv.ParseInt(channelId, 10, 64) // TODO
	return &Client{
		httpClient: httpClient,
		botToken:   botToken,
		channelId:  chatIdInt,
	}
}
