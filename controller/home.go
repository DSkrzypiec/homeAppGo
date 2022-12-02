package controller

import (
	"fmt"
	"homeApp/auth"
	"homeApp/db"
	"homeApp/front"
	"net/http"

	"github.com/rs/zerolog/log"
)

const contrHomePrefix = "controller/home"

type Home struct {
	DbClient    *db.Client
	UserAuth    auth.UserAuthenticator
	AppVersion  string
	CurrentHash string
}

type HomeSummary struct {
	SessionExpUnix           int64 // session expiration timestamp in UNIX millis
	LoadingError             *string
	ColdWaterWeeklyAvg       string // float64 rounded to 2 decimal places
	HotWaterWeeklyAvg        string // float64 rounded to 2 decimal places
	EnergyWeeklyAvg          string // float64 rounded to 2 decimal places
	DocumentsNumber          int
	DocumentsSizeMb          string // float64 rounded to 2 decimal places
	DocumentLatestUploadDate string
	EbooksNumber             int
	EbooksSizeMb             string // float64 rounded to 2 decimal places
	EbookLatestUploadDate    string
	BooksNumber              int
	BookLatestUploadDate     string
	FinancialLatestOrderDate string

	AppVersion  string
	CurrentHash string
}

func (h *Home) HomeSummaryView(w http.ResponseWriter, r *http.Request) {
	tmpl := front.Home()

	sessCookie, cookieErr := r.Cookie(auth.SessCookieName)
	if cookieErr != nil {
		log.Error().Err(cookieErr).Msgf("[%s] cannot get session cookie", contrHomePrefix)
		displayError := "cannot get session Cookie"
		tmpl.Execute(w, HomeSummary{LoadingError: &displayError})
		return
	}

	tokenStatus, tErr := h.UserAuth.IsJwtTokenValid(sessCookie.Value)
	if tErr != nil {
		log.Error().Err(cookieErr).Msgf("[%s] error while validating session cookie", contrHomePrefix)
		displayError := "error while validating session cookie"
		tmpl.Execute(w, HomeSummary{LoadingError: &displayError})
		return
	}

	dbSummary, dbErr := h.DbClient.HomeSummary()
	if dbErr != nil {
		log.Error().Err(dbErr).Msgf("[%s] cannot read home summary from database", contrHomePrefix)
		displayError := "could not read summary from the database"
		tmpl.Execute(w, HomeSummary{LoadingError: &displayError})
		return
	}

	summary := HomeSummary{
		SessionExpUnix:           tokenStatus.TokenExpUnix,
		ColdWaterWeeklyAvg:       fmt.Sprintf("%.2f", dbSummary.ColdWaterWeeklyAvg),
		HotWaterWeeklyAvg:        fmt.Sprintf("%.2f", dbSummary.HotWaterWeeklyAvg),
		EnergyWeeklyAvg:          fmt.Sprintf("%.2f", dbSummary.EnergyWeeklyAvg),
		DocumentsNumber:          dbSummary.DocumentsNumber,
		DocumentsSizeMb:          fmt.Sprintf("%.2f", dbSummary.DocumentsSizeMb),
		DocumentLatestUploadDate: dbSummary.DocumentLatestUploadDate,
		EbooksNumber:             dbSummary.EbooksNumber,
		EbooksSizeMb:             fmt.Sprintf("%.2f", dbSummary.EbooksSizeMb),
		EbookLatestUploadDate:    dbSummary.EbookLatestUploadDate,
		BooksNumber:              dbSummary.BooksNumber,
		BookLatestUploadDate:     dbSummary.BookLatestUploadDate,
		FinancialLatestOrderDate: dbSummary.FinancialLatestOrderDate,

		AppVersion:  h.AppVersion,
		CurrentHash: h.CurrentHash,
	}
	tmpl.Execute(w, summary)
}
