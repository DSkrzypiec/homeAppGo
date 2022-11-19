package main

import (
	"fmt"
	"net/http"

	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/controller"
	"homeApp/db"
	"homeApp/monitor"

	"github.com/rs/zerolog/log"

	_ "modernc.org/sqlite"
)

const (
	SessCookieName = "session"
)

func main() {
	config := ParseConfigFlags()
	setupZerolog()

	dbClient, dbErr := db.NewClient(fmt.Sprintf("file:%s?cache=shared&mode=rw", config.DatabasePath))
	if dbErr != nil {
		log.Fatal().Err(dbErr).Msg("Cannot connect to SQLite")
	}

	// Long-lived
	httpClient := http.Client{Timeout: config.HttpClientTimeout}
	var telegramClient *telegram.Client = nil
	var monitoringMsgSender monitor.MessageSender = monitor.MockMessageSender{}
	if config.UseTelegram2FA {
		telegramClient = telegram.NewClient(&httpClient, config.Telegram.BotToken, config.Telegram.ChannelId)
		monitoringMsgSender = telegramClient
	}

	pageViews := monitor.NewPageViews(config.UseTelegram2FA)
	go pageViews.PublishViews(config.PublishViewsAfter, monitoringMsgSender)

	userAuth := auth.UserAuth{
		DbClient:       dbClient,
		JwtSigningKey:  []byte("crap"), // TODO randomly generate key on each program start
		JwtExpMinutes:  config.SessionTimeoutMinutes,
		TelegramClient: telegramClient,
	}
	authHandlerMan := auth.HandlerManager{UserAuthenticator: userAuth}
	homeContr := controller.Home{
		DbClient:    dbClient,
		AppVersion:  config.AppVersion,
		CurrentHash: config.CurrentCommitSHA,
	}
	counterContr := controller.Counters{
		DbClient:       dbClient,
		TelegramClient: telegramClient,
		UserAuth:       userAuth,
	}
	documentsContr := controller.Documents{
		DbClient:       dbClient,
		TelegramClient: telegramClient,
		UserAuth:       userAuth,
	}
	finContr := controller.Finance{
		DbClient:       dbClient,
		TelegramClient: telegramClient,
		UserAuth:       userAuth,
	}
	loginContr := controller.LoginForm{
		TelegramClient: telegramClient,
		AuthManager:    authHandlerMan,
		DbClient:       dbClient,
		AppVersion:     config.AppVersion,
		CurrentHash:    config.CurrentCommitSHA,
	}

	http.HandleFunc("/", pageViews.Listen(loginContr.LoginFormHandler))
	http.HandleFunc("/login", pageViews.Listen(authHandlerMan.Login))
	http.HandleFunc("/home", pageViews.Listen(authHandlerMan.CheckAuth(homeContr.HomeSummaryView)))
	http.HandleFunc("/counters", pageViews.Listen(authHandlerMan.CheckAuth(counterContr.CountersViewHandler)))
	http.HandleFunc("/counters-new", pageViews.Listen(authHandlerMan.CheckAuth(counterContr.CountersInsertForm)))
	http.HandleFunc("/counters/upload", pageViews.Listen(authHandlerMan.CheckAuth(counterContr.CountersUploadNew)))
	http.HandleFunc("/documents", pageViews.Listen(authHandlerMan.CheckAuth(documentsContr.DocumentsViewHandler)))
	http.HandleFunc("/documents-new", pageViews.Listen(authHandlerMan.CheckAuth(documentsContr.DocumentsInsertForm)))
	http.HandleFunc("/documents/uploadFile", pageViews.Listen(authHandlerMan.CheckAuth(documentsContr.InsertNewDocument)))
	http.HandleFunc("/documentFile", pageViews.Listen(authHandlerMan.CheckAuth(documentsContr.PreviewDocument)))
	http.HandleFunc("/finance", pageViews.Listen(authHandlerMan.CheckAuth(finContr.FinanceViewHandler)))
	http.HandleFunc("/finance-new", pageViews.Listen(authHandlerMan.CheckAuth(finContr.FinanceInsertForm)))
	http.HandleFunc("/finance/upload", pageViews.Listen(authHandlerMan.CheckAuth(finContr.FinanceUploadFile)))
	http.HandleFunc("/logout", pageViews.Listen(authHandlerMan.TerminateSession))

	log.Info().Msgf("Listening on :%d...", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
