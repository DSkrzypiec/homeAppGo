package controller

import (
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/front"
	"net/http"

	"github.com/rs/zerolog/log"
)

const loginFormPrefix = "controller/loginForm"

type LoginForm struct {
	TelegramClient *telegram.Client
	DbClient       *db.Client
	AuthManager    auth.HandlerManager
	AppVersion     string
	CurrentHash    string
}

func (lf *LoginForm) LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	sessionCookieValid, err := lf.AuthManager.IsSessionCookieValid(r)
	if err != nil || !sessionCookieValid {
		log.Info().Msgf("[%s] no session cookie or invalid, rendering login form", loginFormPrefix)
		tmpl := front.Login()
		tmpl.Execute(w, lf)
		return
	}
	homeController := Home{
		DbClient:    lf.DbClient,
		UserAuth:    lf.AuthManager.UserAuthenticator,
		AppVersion:  lf.AppVersion,
		CurrentHash: lf.CurrentHash,
	}
	homeController.HomeSummaryView(w, r)
}
