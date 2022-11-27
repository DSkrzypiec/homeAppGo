package controller

import (
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/front"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	contrBookPrefix = "controller/books"
)

type Books struct {
	DbClient       *db.Client
	TelegramClient *telegram.Client
	UserAuth       auth.UserAuthenticator
}

func (b *Books) BooksViewHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := front.Books()

	execErr := tmpl.Execute(w, nil) // TODO Books from database
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render books view", contrBookPrefix)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func (b *Books) InsertNewBook(w http.ResponseWriter, r *http.Request) {
	tmpl := front.BooksNewForm()
	execErr := tmpl.Execute(w, nil)
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render book insert form", contrBookPrefix)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
	}
}
