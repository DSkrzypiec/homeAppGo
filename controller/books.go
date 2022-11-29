package controller

import (
	"bytes"
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/front"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	contrBookPrefix = "controller/books"
	maxBookSize     = int64(100 * 1024 * 1024) // 100 MiB
)

type Books struct {
	DbClient       *db.Client
	TelegramClient *telegram.Client
	UserAuth       auth.UserAuthenticator
}

type BookList struct {
	Books []db.Book
	Error *string
}

func (b *Books) BooksViewHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := front.Books()

	books, bErr := b.DbClient.Books()
	if bErr != nil {
		log.Error().Err(bErr).Msgf("[%s] cannot load books from database", contrBookPrefix)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// TODO: handle link to e-book file if it was provided otherwise just a
	// title string
	execErr := tmpl.Execute(w, BookList{Books: books}) // TODO Books from database
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render books view", contrBookPrefix)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func (b *Books) BooksInsertForm(w http.ResponseWriter, r *http.Request) {
	tmpl := front.BooksNewForm()

	execErr := tmpl.Execute(w, nil)
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render book insert form", contrBookPrefix)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
	}
}

func (b *Books) InsertNewBook(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start parsing new book form", contrBookPrefix)
	tmpl := front.Books()

	r.ParseMultipartForm(maxBookSize)
	var buf bytes.Buffer
	file, _, fErr := r.FormFile("bookFile")
	if fErr == nil {
		io.Copy(&buf, file)
		defer file.Close()
	}

	newBook := formValuesToNewBook(r, buf.Bytes())

	dErr := b.DbClient.BookInsertNew(newBook)
	if dErr != nil {
		log.Error().Err(dErr).Msgf("[%s] uploading new book failed", contrBookPrefix)
		msg := "Could not insert new book"
		tmpl.Execute(w, BookList{Error: &msg})
		return
	}

	log.Info().Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished parsing and saving new book", contrBookPrefix)

	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

func formValuesToNewBook(r *http.Request, parsedBookFile []byte) db.NewBook {
	titleFv := r.FormValue("title")
	authorsFv := r.FormValue("authors")
	publisherFv := r.FormValue("publisher")
	publishingYearFv := r.FormValue("publishingYear")
	fileExtFv := r.FormValue("fileExt")

	var publisher *string
	if len(publisherFv) > 0 {
		publisher = &publisherFv
	}

	var publishingYear *int
	if len(publishingYearFv) > 0 {
		yearInt, parseErr := strconv.ParseInt(publishingYearFv, 10, 32)
		if parseErr == nil {
			yearInt32 := int(yearInt)
			publishingYear = &yearInt32
		}
	}

	var fileExt *string
	var fileSize *int
	if len(parsedBookFile) > 0 {
		fileExt = &fileExtFv
		size := len(parsedBookFile)
		fileSize = &size
	}

	return db.NewBook{
		Title:          titleFv,
		Authors:        authorsFv,
		Publisher:      publisher,
		PublishingYear: publishingYear,
		FileExtension:  fileExt,
		FileSize:       fileSize,
		BookFile:       parsedBookFile,
	}
}
