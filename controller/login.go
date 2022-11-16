package controller

import (
	"homeApp/auth"
	"homeApp/db"
	"homeApp/front"
	"net/http"

	"github.com/rs/zerolog/log"
)

const loginFormPrefix = "controller/loginForm"

type LoginForm struct {
	DbClient    *db.Client
	AuthManager auth.HandlerManager
}

func (lf *LoginForm) LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	sessionCookieValid, err := lf.AuthManager.IsSessionCookieValid(r)
	if err != nil || !sessionCookieValid {
		log.Info().Msgf("[%s] no session cookie or invalid, rendering login form", loginFormPrefix)
		tmpl := front.Login()
		tmpl.Execute(w, nil)
		return
	}

	homeController := Home{DbClient: lf.DbClient}
	homeController.HomeSummaryView(w, r)
}
