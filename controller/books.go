package controller

import (
	"bytes"
	"fmt"
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/front"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
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

type Book struct {
	Id             int
	Title          template.HTML // string title or link to ebook
	Authors        string
	Publisher      *string
	PublishingYear *int
	Category       string
	Language       string
	FileExtension  *string
	FileSize       *int
}

type BookList struct {
	Books []Book
	Error *string
}

func (b *Books) BooksViewHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filterPhrase := r.FormValue("bookFilter")
	tmpl := front.Books()

	var books []db.Book
	var bErr error

	if filterPhrase == "" {
		books, bErr = b.DbClient.Books()
		if bErr != nil {
			log.Error().Err(bErr).Msgf("[%s] cannot load books from database", contrBookPrefix)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	} else {
		log.Info().Str("filter", filterPhrase).Msgf("[%s] loading books in filtered version", contrBookPrefix)
		books, bErr = b.DbClient.BooksFiltered(filterPhrase)
		if bErr != nil {
			log.Error().Err(bErr).Msgf("[%s] cannot load filtered books from database", contrBookPrefix)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	}

	execErr := tmpl.Execute(w, BookList{Books: booksToDisplay(books)})
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

	var msg string
	if buf.Len() > 0 {
		msg = fmt.Sprintf("Uploaded new e-book [%s] by [%s] of size %.2f MB in [%s] category.",
			newBook.Title, newBook.Authors, float64(buf.Len())/1000000.0, newBook.Category)
	} else {
		msg = fmt.Sprintf("Uploaded new book info [%s] by [%s] in [%s] category.",
			newBook.Title, newBook.Authors, newBook.Category)
	}
	teleErr := SendTelegramMsgForUser(r, b.UserAuth, b.TelegramClient, b.DbClient, msg)
	if teleErr != nil {
		log.Error().Err(teleErr).Msgf("[%s] sending message to Telegram failed", contrFinPrefix)
	}

	log.Info().Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished parsing and saving new book", contrBookPrefix)

	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

// Loads book content from database and send it into ResponseWriter.
func (b *Books) DownloadBook(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start preparing book preview", contrBookPrefix)

	bookId := r.URL.Query().Get("id")
	if bookId == "" {
		log.Error().Msgf("[%s] there is no bookId 'id' URL parameter", contrBookPrefix)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}
	docId, convErr := strconv.ParseInt(bookId, 10, 32)
	if convErr != nil {
		log.Error().Str("id", bookId).Msgf("[%s] cannot convert URL parameter 'id' to int", contrBookPrefix)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}
	docFile, title, fileExt, dbErr := b.DbClient.BookFile(int(docId))
	if dbErr != nil {
		log.Error().Err(dbErr).Str("id", bookId).
			Msgf("[%s] cannot load book content from database", contrBookPrefix)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	newTitle := strings.ReplaceAll(title, " ", "_")
	w.Header().Set("Content-Disposition", "attachment; filename="+newTitle+"."+fileExt)
	w.Header().Set("Content-type", fileExtToContentType(fileExt))
	buff := bytes.NewBuffer(docFile)
	bytesWritten, writeErr := buff.WriteTo(w)
	if writeErr != nil {
		log.Error().Err(writeErr).Str("id", bookId).
			Msgf("[%s] cannot write book content on the response writer", contrBookPrefix)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	if bytesWritten != int64(len(docFile)) {
		log.Error().Str("id", bookId).Int64("bytesWritten", bytesWritten).
			Int("bytesExpected", len(docFile)).Msgf("[%s] incorrect number of bytes written", contrBookPrefix)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	log.Info().Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished preparing book preview", contrBookPrefix)
}

func formValuesToNewBook(r *http.Request, parsedBookFile []byte) db.NewBook {
	category := r.FormValue("category")
	titleFv := r.FormValue("title")
	authorsFv := r.FormValue("authors")
	publisherFv := r.FormValue("publisher")
	publishingYearFv := r.FormValue("publishingYear")
	language := r.FormValue("language")
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
		Category:       category,
		Language:       language,
		FileExtension:  fileExt,
		FileSize:       fileSize,
		BookFile:       parsedBookFile,
	}
}

func booksToDisplay(books []db.Book) []Book {
	bs := make([]Book, len(books))

	for id, book := range books {
		title := template.HTML(book.Title)
		if book.FileSize != nil && *book.FileSize > 0 {
			title = template.HTML(prepBookFileUrl(book))
		}

		bs[id] = Book{
			Id:             book.Id,
			Title:          title,
			Authors:        book.Authors,
			Publisher:      book.Publisher,
			PublishingYear: book.PublishingYear,
			Category:       book.Category,
			Language:       book.Language,
			FileExtension:  book.FileExtension,
			FileSize:       book.FileSize,
		}
	}

	return bs
}

func prepBookFileUrl(book db.Book) string {
	return fmt.Sprintf(`<a href="/bookFile?id=%d" target="_blank">%s</a>`,
		book.Id, book.Title)
}
