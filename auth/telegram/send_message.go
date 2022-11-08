package telegram

import (
	"fmt"
)

const teleSendPrefix = "telegram/sendMsg"

// SendMessage sends a text message onto configured Telegram channel in the
// Client. In case when sending message failed then non-nil error would be
// returned.
func (c *Client) SendMessage(text string) error {
	_, err := c.getRequest(c.sendMessageUrl(text), "sendMessage")
	return err
}

func (c *Client) sendMessageUrl(text string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%d&text=%s",
		c.botToken, c.channelId, text)
}
