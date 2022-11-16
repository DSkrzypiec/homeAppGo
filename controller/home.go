package controller

import (
	"fmt"
	"homeApp/db"
	"homeApp/front"
	"net/http"

	"github.com/rs/zerolog/log"
)

const contrHomePrefix = "controller/home"

type Home struct {
	DbClient *db.Client
}

type HomeSummary struct {
	LoadingError             *string
	ColdWaterWeeklyAvg       string // float64 rounded to 2 decimal places
	HotWaterWeeklyAvg        string // float64 rounded to 2 decimal places
	EnergyWeeklyAvg          string // float64 rounded to 2 decimal places
	DocumentsNumber          int
	DocumentsSizeMb          string // float64 rounded to 2 decimal places
	DocumentLatestUploadDate string
	FinancialLatestOrderDate string
}

func (h *Home) HomeSummaryView(w http.ResponseWriter, _ *http.Request) {
	tmpl := front.Home()
	dbSummary, dbErr := h.DbClient.HomeSummary()
	if dbErr != nil {
		log.Error().Err(dbErr).Msgf("[%s] cannot read home summary from database", contrHomePrefix)
		displayError := "could not read summary from the database"
		tmpl.Execute(w, HomeSummary{LoadingError: &displayError})
		return
	}

	summary := HomeSummary{
		ColdWaterWeeklyAvg:       fmt.Sprintf("%.2f", dbSummary.ColdWaterWeeklyAvg),
		HotWaterWeeklyAvg:        fmt.Sprintf("%.2f", dbSummary.HotWaterWeeklyAvg),
		EnergyWeeklyAvg:          fmt.Sprintf("%.2f", dbSummary.EnergyWeeklyAvg),
		DocumentsNumber:          dbSummary.DocumentsNumber,
		DocumentsSizeMb:          fmt.Sprintf("%.2f", dbSummary.DocumentsSizeMb),
		DocumentLatestUploadDate: dbSummary.DocumentLatestUploadDate,
		FinancialLatestOrderDate: dbSummary.FinancialLatestOrderDate,
	}
	tmpl.Execute(w, summary)
}
