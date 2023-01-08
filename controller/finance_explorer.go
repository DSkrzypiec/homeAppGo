package controller

import (
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/front"
	"net/http"
)

const (
	contrFinExPrefix = "controller/finEx"
)

type FinanceExplorer struct {
	DbClient       *db.Client
	TelegramClient *telegram.Client
	UserAuth       auth.UserAuthenticator
}

// TODO
func (fw *FinanceExplorer) FinanceExplorerViewHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := front.FinanceExplorer()
	tmpl.Execute(w, nil)
}
