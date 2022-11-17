package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/front"

	"github.com/rs/zerolog/log"
)

const (
	contrDocPrefix  = "controller/docs"
	maxDocumentSize = int64(100 * 1024 * 1024) // 100 MiB
)

type Documents struct {
	DbClient       *db.Client
	TelegramClient *telegram.Client
	UserAuth       auth.UserAuthenticator
}

type DocumentsList struct {
	Documents []db.DocumentInfo
}

func (d *Documents) DocumentsViewHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filterDocPhrase := r.FormValue("docFilter")

	var documents []db.DocumentInfo
	var dErr error

	if filterDocPhrase == "" {
		documents, dErr = d.DbClient.Documents()
		if dErr != nil {
			log.Error().Err(dErr).Msgf("[%s] cannot load documents from database", contrDocPrefix)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	} else {
		log.Info().Str("filter", filterDocPhrase).Msgf("[%s] loading documents in filtered version", contrDocPrefix)
		documents, dErr = d.DbClient.DocumentsFiltered(filterDocPhrase)
		if dErr != nil {
			log.Error().Err(dErr).Msgf("[%s] cannot load filtered documents from database", contrDocPrefix)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	}

	tmpl := front.Documents()
	execErr := tmpl.Execute(w, DocumentsList{Documents: documents})
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render document list view", contrDocPrefix)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
}

func (d *Documents) DocumentsInsertForm(w http.ResponseWriter, r *http.Request) {
	tmpl := front.DocumentsNewForm()
	execErr := tmpl.Execute(w, nil)
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render document insert form", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
	}
}

// Inserts new document. Both document meta information and its content.
func (d *Documents) InsertNewDocument(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start parsing new document form", contrDocPrefix)

	r.ParseMultipartForm(maxDocumentSize)
	var buf bytes.Buffer
	file, _, err := r.FormFile("docFile")
	if err != nil {
		log.Error().Err(err).Msgf("[%s] couldn't get file from the form", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
		return
	}
	defer file.Close()
	io.Copy(&buf, file)

	docName := r.FormValue("documentName")
	docDate := r.FormValue("documentDate")
	docCategory := r.FormValue("category")
	docPerson := r.FormValue("person")
	docExt := r.FormValue("fileExt")

	newDoc := db.NewDocument{
		Name:           docName,
		DocumentDate:   docDate,
		Category:       docCategory,
		PersonInvolved: docPerson,
		FileExtension:  docExt,
		DocumentFile:   buf.Bytes(),
	}

	dErr := d.DbClient.DocumentInsertNew(newDoc)
	if dErr != nil {
		log.Error().Err(dErr).Msgf("[%s] uploading new document failed", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
		return
	}

	msg := fmt.Sprintf("Uploaded new document [%s][%s] of size %.2f MB", docName, docExt,
		float64(buf.Len())/1000000.0)
	teleErr := SendTelegramMsgForUser(r, d.UserAuth, d.TelegramClient, d.DbClient, msg)
	if teleErr != nil {
		log.Error().Err(teleErr).Msgf("[%s] sending message to Telegram failed", contrFinPrefix)
	}

	log.Info().Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished parsing and saving new document", contrDocPrefix)

	http.Redirect(w, r, "/documents", http.StatusSeeOther)
}

// Loads document content from database and send it into ResponseWriter.
func (d *Documents) PreviewDocument(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start preparing document preview", contrDocPrefix)

	documentId := r.URL.Query().Get("id")
	if documentId == "" {
		log.Error().Msgf("[%s] there is no documentId 'id' URL parameter", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
		return
	}
	docId, convErr := strconv.ParseInt(documentId, 10, 32)
	if convErr != nil {
		log.Error().Str("id", documentId).Msgf("[%s] cannot convert URL parameter 'id' to int", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
		return
	}
	docFile, fileExt, dbErr := d.DbClient.DocumentFile(int(docId))
	if dbErr != nil {
		log.Error().Err(dbErr).Str("id", documentId).
			Msgf("[%s] cannot load document content from database", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-type", fileExtToContentType(fileExt))
	buff := bytes.NewBuffer(docFile)
	bytesWritten, writeErr := buff.WriteTo(w)
	if writeErr != nil {
		log.Error().Err(writeErr).Str("id", documentId).
			Msgf("[%s] cannot write document content on the response writer", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
		return
	}

	if bytesWritten != int64(len(docFile)) {
		log.Error().Str("id", documentId).Int64("bytesWritten", bytesWritten).
			Int("bytesExpected", len(docFile)).Msgf("[%s] incorrect number of bytes written", contrDocPrefix)
		http.Redirect(w, r, "/documents", http.StatusSeeOther)
		return
	}

	log.Info().Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished preparing document preview", contrDocPrefix)
}

func fileExtToContentType(fileExt string) string {
	fe := strings.ToLower(fileExt)

	switch fe {
	case "pdf":
		return "application/pdf"
	case "mp4":
		return "video/mp4"
	case "gif":
		return "image/gif"
	case "png":
		return "image/png"
	case "jpeg":
		return "image/jpeg"
	case "jpg":
		return "image/jpeg"
	case "zip":
		return "application/zip"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "doc":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "text/plain"
	}
}
