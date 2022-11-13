package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	teleUpdatesPrefix          = "telegram/updates"
	updatesLimit               = 25
	updatesQueryTimeoutSeconds = 10
	maxNumberOfRetries         = 20
	secondsBeforeRetry         = 5
)

var ErrTelegramUser2FATimeout = errors.New("timeout, there wasn't correct user input on Telegram channel")

// CheckMessageWithPattern periodically (by secondsBeforeRetry) reads Telegram
// updates (last updatesLimit messages from the Telegram channel) and check if
// given pattern occurs within text of those updates. Messages only from
// specific channel (c.channelId) are checked. Updates are loaded until either
// success or "twoFaTimeout" timeout.
func (c *Client) CheckMessageWithPattern(pattern string, twoFaTimeout time.Duration) (bool, error) {
	log.Info().Msgf("[%s] start checking telegram chat messages", teleUpdatesPrefix)
	startTs := time.Now()
	startTsUnixSeconds := int(startTs.UnixMilli() / 1000)
	isMatchedChan := make(chan bool)
	errChan := make(chan error)
	timeout := time.After(twoFaTimeout)

	lastMessageUpdateId := c.getLastUpdateId()
	go c.getUpdatesWithPattern(startTsUnixSeconds, lastMessageUpdateId, pattern, isMatchedChan, errChan)

	for {
		select {
		case <-timeout:
			// User 2FA action timeouted
			log.Error().Err(ErrTelegramUser2FATimeout).Dur("duration", time.Since(startTs)).
				Msgf("[%s] timeout", teleUpdatesPrefix)
			c.SendMessage(fmt.Sprintf("[2FA] Timeout! I didn't receive input [%s] from user. Please try again.", pattern))
			return false, ErrTelegramUser2FATimeout

		case err := <-errChan:
			// An error during single Telegram communication, retrying after a pause
			log.Error().Err(err).Msgf("[%s] error while checking Telegram updates", teleUpdatesPrefix)
			time.Sleep(secondsBeforeRetry * time.Second)
			go c.getUpdatesWithPattern(startTsUnixSeconds, lastMessageUpdateId, pattern, isMatchedChan, errChan)

		case isMatched := <-isMatchedChan:
			// got result, if pattern is matched true is returned else we retry until timeout or other success
			if isMatched {
				log.Info().Str("pattern", pattern).Msgf("[%s] found matching pattern in Telegram updates", teleUpdatesPrefix)
				c.SendMessage(fmt.Sprintf("[2FA] [%s] confirmed!", pattern))
				return true, nil
			} else {
				log.Info().Str("pattern", pattern).
					Msgf("[%s] parsed Telegram updates but pattern was not matched", teleUpdatesPrefix)
				time.Sleep(secondsBeforeRetry * time.Second)
				go c.getUpdatesWithPattern(startTsUnixSeconds, lastMessageUpdateId, pattern, isMatchedChan, errChan)
			}
		}
	}
}

// This method gets recent messages from the Telegram channel and try to find
// message with matching pattern. Either errors or results are sent over
// channels (Go channels :)).
func (c *Client) getUpdatesWithPattern(startTsUnixSeconds int, lastMessageUpdateId *int, pattern string,
	isMatchedChan chan<- bool, errChan chan error) {

	updatesUrl := c.getUpdatesUrl(updatesLimit, updatesQueryTimeoutSeconds, lastMessageUpdateId)
	apiResp, reqErr := c.getRequest(updatesUrl, "getUpdates")
	if reqErr != nil {
		errChan <- reqErr
		return
	}

	var updates []Update
	jErr := json.Unmarshal(apiResp.Result, &updates)
	if jErr != nil {
		errChan <- jErr
		return
	}

	isMatchedChan <- matchPatternWithinMessageText(startTsUnixSeconds, pattern, c.channelId, updates)
}

// This function tries to match given pattern to Telegram chat post message
// text which happens later than startTsUnixSeconds on particular chat (based
// on chatId).
func matchPatternWithinMessageText(startTsUnixSeconds int, pattern string, chatId int64, messages []Update) bool {
	for _, update := range messages {
		if update.ChannelPost == nil {
			log.Warn().Msgf("[%s] empty channel post", teleUpdatesPrefix)
			continue
		}
		if update.ChannelPost.Chat == nil {
			log.Warn().Msgf("[%s] empty channel post chat", teleUpdatesPrefix)
			continue
		}
		if update.ChannelPost.Date <= startTsUnixSeconds || update.ChannelPost.Chat.ID != chatId {
			continue
		}
		log.Info().Str("text", update.ChannelPost.Text).Msgf("[%s] text", teleUpdatesPrefix)
		if strings.Contains(strings.ToLower(update.ChannelPost.Text), strings.ToLower(pattern)) {
			log.Info().Str("pattern", pattern).Str("telegramText", update.ChannelPost.Text).
				Msgf("[%s] found match", teleUpdatesPrefix)
			return true
		}
	}
	return false
}

// Gets Update ID of last message on the Telegram channel. In case of error nil
// would be returned and handled by callee.
func (c *Client) getLastUpdateId() *int {
	errorDuringLastMsg := false
	lastMessageResp, lastReqErr := c.getRequest(c.getUpdateLastUrl(), "getUpdateLast")
	if lastReqErr != nil {
		log.Error().Err(lastReqErr).Msgf("[%s] failed to getUpdate for last message", teleUpdatesPrefix)
		errorDuringLastMsg = true

	}
	var lastMessage []Update
	lmJsonErr := json.Unmarshal(lastMessageResp.Result, &lastMessage)
	if lmJsonErr != nil {
		log.Error().Err(lmJsonErr).Msgf("[%s] failed JSON unmarshal last message", teleUpdatesPrefix)
		errorDuringLastMsg = true
	}

	if errorDuringLastMsg {
		return nil
	}

	if len(lastMessage) == 0 {
		log.Info().Msgf("[%s] unmarshaled last update but got zero updates", teleUpdatesPrefix)
		return nil
	}

	log.Info().Int("lastUpdateId", lastMessage[0].UpdateID).Msgf("[%s] got last update ID", teleUpdatesPrefix)
	return &lastMessage[0].UpdateID
}

func (c *Client) getUpdatesUrl(limit, timeout int, offset *int) string {
	if offset != nil {
		return fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?limit=%d&timeout=%d&offset=%d",
			c.botToken, limit, timeout, *offset)
	}
	return fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?limit=%d&timeout=%d",
		c.botToken, limit, timeout)
}

func (c *Client) getUpdateLastUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=-1", c.botToken)
}
