package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const teleGetPrefix = "telegram/get"

// Generic GET request for Telegram API. Handles errors and returns
// APIResponse.
func (c *Client) getRequest(telegramUrl, endpointName string) (APIResponse, error) {
	startTs := time.Now()
	var apiResp APIResponse

	log.Info().Str("endpoint", endpointName).Msgf("[%s] start sending GET request", teleGetPrefix)

	resp, err := c.httpClient.Get(telegramUrl)
	if err != nil {
		log.Error().Err(err).Dur("duration", time.Since(startTs)).
			Msgf("[%s] telegram GET [%s] failed", teleGetPrefix, endpointName)
		return apiResp, err
	}

	body, rErr := ioutil.ReadAll(resp.Body)
	if rErr != nil {
		log.Error().Err(rErr).Dur("duration", time.Since(startTs)).
			Msgf("[%s] couldn't read [%s] response body", teleGetPrefix, endpointName)
		return apiResp, fmt.Errorf("couldn't read response body: %s", rErr.Error())
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("statuscode", resp.StatusCode).
			Str("respBody", string(body)).Dur("duration", time.Since(startTs)).
			Msgf("[%s] got status code != 200 on [%s] response", teleGetPrefix, endpointName)
		return apiResp, fmt.Errorf("got %d status code in [%s] response", resp.StatusCode, endpointName)
	}

	jErr := json.Unmarshal(body, &apiResp)
	if jErr != nil {
		log.Error().Err(jErr).Str("respBody", string(body)).Dur("duration", time.Since(startTs)).
			Msgf("[%s] unmarshal into APIResponse failed", teleGetPrefix)
		return apiResp, fmt.Errorf("couldn't unmarshal into telegram APIResponse: %s", jErr.Error())
	}

	if !apiResp.Ok {
		log.Error().Str("respBody", string(body)).Dur("duration", time.Since(startTs)).
			Msgf("[%s] got and unmarshaled [%s] response but status is not successful", teleGetPrefix, endpointName)
	}

	log.Info().Str("endpoint", endpointName).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished GET request", teleGetPrefix)

	return apiResp, nil
}
