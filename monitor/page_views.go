package monitor

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	MaxRetriesWithoutMessage      = 5
	MaxSendMessageRetries         = 10
	WaitBetweenSendMessageRetries = 60 * time.Second
)

// MessageSender would be used to send monitoring messages.
type MessageSender interface {
	SendMessage(string) error
}

// Trivial MessageSender implementation using logger.
type MockMessageSender struct{}

func (mms MockMessageSender) SendMessage(msg string) error {
	log.Info().Msgf("[monitor/msgsender] %s", msg)
	return nil
}

// PageViews keeps map with endpoints paths and visit counts.
type PageViews struct {
	sync.Mutex
	urlCounts          map[string]int64
	escapeMessageToUrl bool
}

// NewPageViews creates new PageViews.
func NewPageViews(escapeMessageToUrl bool) *PageViews {
	return &PageViews{
		urlCounts:          map[string]int64{},
		escapeMessageToUrl: escapeMessageToUrl,
	}
}

// Listen is a middleware which logs URL that have been visited and call next HTTP handler.
func (pv *PageViews) Listen(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" {
			pv.addView("/")
		} else {
			pv.addView(r.URL.EscapedPath())
		}
		next(w, r)
	}
}

// PublishViews after each period time reports page views statistics using
// messenger. If there is no statistics in given period, then message will no
// be sent. In case when there will no statistics in MaxRetriesWithoutMessage
// periods, then message will be sent regarding last update and information
// about no visits.
func (pv *PageViews) PublishViews(period time.Duration, messenger MessageSender) {
	lastUpdate := time.Now()
	noViewsInPeriods := 0
	for {
		time.Sleep(period)
		if len(pv.urlCounts) == 0 && noViewsInPeriods < MaxRetriesWithoutMessage {
			noViewsInPeriods = noViewsInPeriods + 1
			continue
		}
		if len(pv.urlCounts) == 0 && noViewsInPeriods >= MaxRetriesWithoutMessage {
			msg := fmt.Sprintf("[Info] No page views since %s.",
				lastUpdate.Format("2006-01-02 15:04:05"))
			if pv.escapeMessageToUrl {
				msg = url.QueryEscape(msg)
			}
			messenger.SendMessage(msg)
			lastUpdate = time.Now()
			noViewsInPeriods = 0
			continue
		}
		msg := fmt.Sprintf("[Info] Page views since [%s]:\n%s",
			lastUpdate.Format("2006-01-02 15:04:05"), pv.toString())
		if pv.escapeMessageToUrl {
			msg = url.QueryEscape(msg)
		}
		messenger.SendMessage(msg)
		pv.urlCounts = map[string]int64{}
		lastUpdate = time.Now()
		noViewsInPeriods = 0
	}
}

func publishMessage(msg string, messenger MessageSender) {
	retryId := 0
	for {
		sendErr := messenger.SendMessage(msg)
		if sendErr == nil {
			return
		}
		if sendErr != nil {
			if retryId >= MaxSendMessageRetries {
				log.Error().Err(sendErr).
					Msgf("[monitor/page_views] after %d retries seng message failed", MaxSendMessageRetries)
				return
			}
			retryId = retryId + 1
			time.Sleep(WaitBetweenSendMessageRetries)
		}
	}
}

func (pv *PageViews) addView(url string) {
	pv.Lock()
	if len(url) > 0 {
		pv.urlCounts[url] = pv.urlCounts[url] + 1
	}
	pv.Unlock()
}

func (pv *PageViews) toString() string {
	var b strings.Builder

	pv.Lock()
	keys := make([]string, 0, len(pv.urlCounts))
	for k := range pv.urlCounts {
		keys = append(keys, k)
	}
	pv.Unlock()

	sort.Strings(keys)

	pv.Lock()
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("  %s: %d\n", k, pv.urlCounts[k]))
	}
	pv.Unlock()

	return b.String()
}
