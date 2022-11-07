package main

import (
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

const SessCookieName = "session"

func main() {
	setupZerolog()

	dbClient, dbErr := db.NewClient("file:test.db?cache=shared&mode=rw")
	if dbErr != nil {
		log.Fatal().Err(dbErr).Msg("Cannot connect to SQLite")
	}

	// Long-lived
	counterContr := controller.Counters{DbClient: dbClient}
	userAuth := auth.UserAuth{
		DbClient:      dbClient,
		JwtSigningKey: []byte("crap"), // TODO randomly generate key on each program start
		JwtExpMinutes: 10,             // TODO
	}
	authHandlerMan := auth.HandlerManager{UserAuthenticator: userAuth}

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
