package controller

import (
	"log"
	"net/http"

	"homeApp/db"
	"homeApp/front"
)

type Counters struct {
	DbClient *db.Client
}

type CountersTemplateInput struct {
	WaterData  []db.WaterCounterEntry
	EnergyData []db.EnergyCounterEntry
}

// TODO:
func (c *Counters) CountersViewHandler(w http.ResponseWriter, _ *http.Request) {
	waterData, wdErr := c.DbClient.WaterData(1, 10)
	if wdErr != nil {
		// TODO
		log.Fatal(wdErr)
	}
	energyData, edErr := c.DbClient.EnergyData(1, 10)
	if wdErr != nil {
		// TODO
		log.Fatal(edErr)
	}

	tmpl := front.Counters()
	tmpl.Execute(w, CountersTemplateInput{waterData, energyData})
}

func (c *Counters) CountersInsertForm(w http.ResponseWriter, _ *http.Request) {
	tmpl := front.CountersNewForm()
	tmpl.Execute(w, nil)
}
