package controller

import (
	"errors"
	"fmt"
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const sendTelPrefix = "controller/sendTelegram"

// Sends message on Telegram channel stating currently authenticated user. If
// Telegram client is not provided (2FA is turned off), then msg is just logged.
func SendTelegramMsgForUser(r *http.Request, userAuth auth.UserAuthenticator, tClient *telegram.Client,
	dbClient *db.Client, msg string) error {
	if tClient == nil {
		log.Info().Msgf("[%s] %s", sendTelPrefix, msg)
		return nil
	}

	startTs := time.Now()
	log.Info().Msgf("[%s] start sending message to Telegram", sendTelPrefix)

	sessCookie, cookieErr := r.Cookie(auth.SessCookieName)
	if cookieErr != nil {
		return cookieErr
	}

	tokenStatus, validErr := userAuth.IsJwtTokenValid(sessCookie.Value)
	if validErr != nil {
		return validErr
	}
	if !tokenStatus.IsValid {
		return errors.New("JWT is invalid, message will not be sent")
	}

	user, uErr := dbClient.UserByUserId(tokenStatus.UserId)
	if uErr != nil {
		return fmt.Errorf("cannot send message, because getting user info failed: %v", uErr)
	}

	teleMsg := fmt.Sprintf("[Info] [%s] %s", user.Username, msg)
	tErr := tClient.SendMessage(teleMsg)
	if tErr != nil {
		return tErr
	}

	log.Info().Dur("duration", time.Since(startTs)).Msgf("[%s] sent message to Telegram", sendTelPrefix)
	return nil
}
