package main

import (
	"net/http"
	"time"

	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/controller"
	"homeApp/db"

	"github.com/rs/zerolog/log"

	_ "modernc.org/sqlite"
)

const (
	SessCookieName = "session"
)

func main() {
	config := ParseConfigFlags()
	setupZerolog()

	dbClient, dbErr := db.NewClient("file:test.db?cache=shared&mode=rw")
	if dbErr != nil {
		log.Fatal().Err(dbErr).Msg("Cannot connect to SQLite")
	}

	// Long-lived
	httpClient := http.Client{Timeout: 60 * time.Second}
	var telegramClient *telegram.Client = nil
	if config.UseTelegram2FA {
		telegramClient = telegram.NewClient(&httpClient, config.Telegram.BotToken, config.Telegram.ChannelId)
	}

	userAuth := auth.UserAuth{
		DbClient:       dbClient,
		JwtSigningKey:  []byte("crap"), // TODO randomly generate key on each program start
		JwtExpMinutes:  10,             // TODO
		TelegramClient: telegramClient,
	}
	authHandlerMan := auth.HandlerManager{UserAuthenticator: userAuth}
	loginContr := controller.LoginForm{AuthManager: authHandlerMan}
	counterContr := controller.Counters{DbClient: dbClient}

	/*
		telegramClient.SendMessage("HomeApp started!")
		r, rErr := telegramClient.CheckMessageWithPattern("damian", 60*time.Second)
		if rErr != nil {
			log.Fatal().Err(rErr)
		}
		log.Info().Bool("matchResult", r).Msg("Got Telegram match")
	*/

	http.HandleFunc("/", loginContr.LoginFormHandler)
	http.HandleFunc("/login", authHandlerMan.Login)
	http.HandleFunc("/home", authHandlerMan.CheckAuth(controller.Home))
	http.HandleFunc("/counters", authHandlerMan.CheckAuth(counterContr.CountersViewHandler))
	http.HandleFunc("/counters-new", authHandlerMan.CheckAuth(counterContr.CountersInsertForm))

	log.Info().Msg("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}
