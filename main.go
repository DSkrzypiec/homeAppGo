package main

import (
	"fmt"
	"net/http"

	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/controller"
	"homeApp/db"
	"homeApp/monitor"
	"homeApp/rand"

	"github.com/rs/zerolog/log"

	_ "modernc.org/sqlite"
)

const (
	SessCookieName   = "session"
	JwtSigningKeyLen = 256
)

func main() {
	config := ParseConfigFlags()
	config.setupZerolog()

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

	registeredEndpoints := make(map[string]struct{}) // To be updated during endpoint registration
	pageViews := monitor.NewPageViews(config.UseTelegram2FA, registeredEndpoints)
	go pageViews.PublishViews(config.PublishViewsAfter, monitoringMsgSender)

	userAuth := auth.UserAuth{
		DbClient:       dbClient,
		JwtSigningKey:  []byte(rand.AlphanumStr(JwtSigningKeyLen)),
		JwtExpMinutes:  config.SessionTimeoutMinutes,
		TelegramClient: telegramClient,
	}
	authHandlerMan := auth.HandlerManager{UserAuthenticator: userAuth}
	homeContr := controller.Home{
		DbClient:    dbClient,
		UserAuth:    userAuth,
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
	booksContr := controller.Books{
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
	sessionContr := controller.Session{
		UserAuth: userAuth,
	}
	endpoints := EndpointRegister{
		PageViews:           pageViews,
		AuthHandler:         &authHandlerMan,
		RegisteredEndpoints: registeredEndpoints,
	}

	endpoints.register("/", loginContr.LoginFormHandler)
	endpoints.register("/login", authHandlerMan.Login)
	endpoints.registerWithAuth("/home", homeContr.HomeSummaryView)
	endpoints.registerWithAuth("/books", booksContr.BooksViewHandler)
	endpoints.registerWithAuth("/books-new", booksContr.BooksInsertForm)
	endpoints.registerWithAuth("/books/upload", booksContr.InsertNewBook)
	endpoints.registerWithAuth("/bookFile", booksContr.DownloadBook)
	endpoints.registerWithAuth("/counters", counterContr.CountersViewHandler)
	endpoints.registerWithAuth("/counters-new", counterContr.CountersInsertForm)
	endpoints.registerWithAuth("/counters/upload", counterContr.CountersUploadNew)
	endpoints.registerWithAuth("/documents", documentsContr.DocumentsViewHandler)
	endpoints.registerWithAuth("/documents-new", documentsContr.DocumentsInsertForm)
	endpoints.registerWithAuth("/documents/uploadFile", documentsContr.InsertNewDocument)
	endpoints.registerWithAuth("/documentFile", documentsContr.PreviewDocument)
	endpoints.registerWithAuth("/finance", finContr.FinanceViewHandler)
	endpoints.registerWithAuth("/finance-new", finContr.FinanceInsertForm)
	endpoints.registerWithAuth("/finance/upload", finContr.FinanceUploadFile)
	endpoints.registerWithAuth("/logout", authHandlerMan.TerminateSession)
	endpoints.registerWithAuth("/session/prolong", sessionContr.ProlongHandler)

	log.Info().Msgf("Listening on :%d...", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}

type EndpointRegister struct {
	PageViews           *monitor.PageViews
	AuthHandler         *auth.HandlerManager
	RegisteredEndpoints map[string]struct{}
}

func (er *EndpointRegister) register(path string, handler func(http.ResponseWriter, *http.Request)) {
	if _, exist := er.RegisteredEndpoints[path]; exist {
		log.Fatal().Msgf("Cannot register more then one endpoint for path [%s]", path)
	}
	er.RegisteredEndpoints[path] = struct{}{}

	// PageViews statistics should be included for every endpoint
	http.HandleFunc(path, er.PageViews.Listen(handler))
}

func (er *EndpointRegister) registerWithAuth(path string, handler func(http.ResponseWriter, *http.Request)) {
	er.register(path, er.AuthHandler.CheckAuth(handler))
}
