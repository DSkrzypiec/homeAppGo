package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/front"

	"github.com/rs/zerolog/log"
)

const contrCountPrefix = "controller/counter"

type Counters struct {
	DbClient       *db.Client
	TelegramClient *telegram.Client
	UserAuth       auth.UserAuthenticator
}

type CountersTemplateInput struct {
	LoadingError *string
	WaterData    []db.WaterCounterEntry
	EnergyData   []db.EnergyCounterEntry
}

type CountersUpload struct {
	UploadError *string
}

// CountersViewHandler serves page with counter data.
func (c *Counters) CountersViewHandler(w http.ResponseWriter, _ *http.Request) {
	startTs := time.Now()
	tmpl := front.Counters()
	log.Info().Msgf("[%s] start reading counters data", contrCountPrefix)

	waterData, wdErr := c.DbClient.WaterData(1, 10) // TODO: paging
	if wdErr != nil {
		log.Error().Err(wdErr).Msgf("[%s] cannot load water counter data", contrCountPrefix)
		displayError := "could not read water counter data from database"
		tmpl.Execute(w, CountersTemplateInput{LoadingError: &displayError})
		return
	}
	energyData, edErr := c.DbClient.EnergyData(1, 10) // TODO: paging
	if edErr != nil {
		log.Error().Err(wdErr).Msgf("[%s] cannot load energy counter data", contrCountPrefix)
		displayError := "could not read energy counter data from database"
		tmpl.Execute(w, CountersTemplateInput{LoadingError: &displayError})
		return
	}

	log.Info().Dur("duration", time.Since(startTs)).Msgf("[%s]", contrCountPrefix)
	tmpl.Execute(w, CountersTemplateInput{nil, waterData, energyData})
}

// CountersInsertForm renders counters insert form page.
func (c *Counters) CountersInsertForm(w http.ResponseWriter, _ *http.Request) {
	tmpl := front.CountersNewForm()
	tmpl.Execute(w, nil)
}

// CountersUploadNew takes counter insert form data and put it into database.
func (c *Counters) CountersUploadNew(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	tmpl := front.CountersNewForm()
	log.Info().Msgf("[%s] start inserting new counters data", contrCountPrefix)

	r.ParseForm()
	date := r.FormValue("inputDate")
	coldWaterLiters := r.FormValue("coldWater")
	hotWaterLiters := r.FormValue("hotWater")
	energy := r.FormValue("energy")

	coldWater, cwErr := strconv.ParseInt(coldWaterLiters, 10, 32)
	hotWater, hwErr := strconv.ParseInt(hotWaterLiters, 10, 32)
	ene, eErr := strconv.ParseFloat(energy, 64)

	if cwErr != nil || hwErr != nil || eErr != nil {
		log.Error().Msgf("[%s] cannot cast to int or float64 inputs from the form", contrCountPrefix)
		displayError := "incorrect input, please try again"
		tmpl.Execute(w, CountersUpload{UploadError: &displayError})
		return
	}

	waterInput := db.WaterCounterEntry{
		Date:            date,
		ColdWaterLiters: int(coldWater),
		HotWaterLiters:  int(hotWater),
	}
	energyInput := db.EnergyCounterEntry{
		Date:      date,
		EnergyKwh: ene,
	}

	dbErr := c.DbClient.CountersInsertNew(waterInput, energyInput)
	if dbErr != nil {
		log.Error().Err(dbErr).Msgf("[%s] cannot insert counters data into database", contrCountPrefix)
		displayError := "could not insert counters data into database"
		tmpl.Execute(w, CountersUpload{UploadError: &displayError})
		return
	}

	msg := fmt.Sprintf("Uploaded new counters state for [%s]: Water{Cold: %d liters, Hot: %d liters}"+
		", Energy %.2f kWh.", date, int(coldWater), int(hotWater), ene)
	teleErr := SendTelegramMsgForUser(r, c.UserAuth, c.TelegramClient, c.DbClient, msg)
	if teleErr != nil {
		log.Error().Err(teleErr).Msgf("[%s] sending message to Telegram failed", contrFinPrefix)
	}

	log.Info().Dur("duration", time.Since(startTs)).Msgf("[%s] finished inserting counters data",
		contrCountPrefix)

	http.Redirect(w, r, "/counters", http.StatusSeeOther)
}
