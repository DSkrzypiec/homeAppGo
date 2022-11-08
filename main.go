package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"homeApp/auth"
	"homeApp/controller"
	"homeApp/db"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "modernc.org/sqlite"
)

const (
	SessCookieName       = "session"
	TelegramBotTokenEnv  = "HOMEAPP_TELEGRAM_BOT_TOKEN"
	TelegramChannelIdEnv = "HOMEAPP_TELEGRAM_CHANNEL_ID"
)

func main() {
	setupZerolog()

	dbClient, dbErr := db.NewClient("file:test.db?cache=shared&mode=rw")
	if dbErr != nil {
		log.Fatal().Err(dbErr).Msg("Cannot connect to SQLite")
	}

	// Telegram info
	telegramBotToken := os.Getenv(TelegramBotTokenEnv)
	telegramChannelId := os.Getenv(TelegramChannelIdEnv)
	fmt.Printf("%s | %s\n", telegramBotToken, telegramChannelId)

	// Long-lived
	// httpClient := http.Client{Timeout: 60 * time.Second}
	// telegramClient := telegram.NewClient(&httpClient, telegramBotToken, telegramChannelId)
	userAuth := auth.UserAuth{
		DbClient:      dbClient,
		JwtSigningKey: []byte("crap"), // TODO randomly generate key on each program start
		JwtExpMinutes: 10,             // TODO
	}
	authHandlerMan := auth.HandlerManager{UserAuthenticator: userAuth}
	counterContr := controller.Counters{DbClient: dbClient}

	/*
		telegramClient.SendMessage("HomeApp started!")
		r, rErr := telegramClient.CheckMessageWithPattern("damian", 60*time.Second)
		if rErr != nil {
			log.Fatal().Err(rErr)
		}
		log.Info().Bool("matchResult", r).Msg("Got Telegram match")
	*/

	http.HandleFunc("/", controller.LoginForm)
	http.HandleFunc("/login", authHandlerMan.Login)
	http.HandleFunc("/home", authHandlerMan.CheckAuth(controller.Home))
	http.HandleFunc("/counters", authHandlerMan.CheckAuth(counterContr.CountersViewHandler))
	http.HandleFunc("/counters-new", authHandlerMan.CheckAuth(counterContr.CountersInsertForm))

	log.Info().Msg("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}

func setupZerolog() {
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.SetGlobalLevel(zerolog.DebugLevel)                                               // TODO: make flag for this
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}) // TODO: make flag for this
}
