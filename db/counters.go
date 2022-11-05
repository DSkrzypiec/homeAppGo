package db

import (
	"time"

	"github.com/rs/zerolog/log"
)

const dbCounterPerfix = "db/counters"

type WaterCounterEntry struct {
	Date            string
	ColdWaterLiters int
	HotWaterLiters  int
}

// WaterData reads water counter entries for specified page and page size with
// descending order by the date of measurement.
func (c *Client) WaterData(page, pageSize int) ([]WaterCounterEntry, error) {
	waterData := make([]WaterCounterEntry, 0, pageSize)

	log.Info().Int("page", page).Int("pageSize", pageSize).
		Msgf("[%s] start reading water counter data", dbCounterPerfix)
	startTs := time.Now()

	rows, qErr := c.dbConn.Query(waterMeasuresQuery(), pageSize, rowsToSkip(page, pageSize))
	if qErr != nil {
		log.Error().Err(qErr).
			Int("page", page).
			Int("pageSize", pageSize).
			Msg("failed waterMeasuresQuery query")
		return waterData, qErr
	}

	var date string
	var coldWater, hotWater int
	for rows.Next() {
		sErr := rows.Scan(&date, &coldWater, &hotWater)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of waterMeasuresQuery", dbCounterPerfix)
			continue
		}
		waterData = append(waterData, WaterCounterEntry{date, coldWater, hotWater})
	}
	elapsed := time.Since(startTs)
	log.Info().Int("rowsLoaded", len(waterData)).Dur("durationMillis", elapsed).
		Msgf("[%s] finished reading water counter data", dbCounterPerfix)

	return waterData, nil
}

type EnergyCounterEntry struct {
	Date      string
	EnergyKwh float64
}

// EnergyData reads energy counter entries for specified page and page size with
// descending order by the date of measurement.
func (c *Client) EnergyData(page, pageSize int) ([]EnergyCounterEntry, error) {
	energyData := make([]EnergyCounterEntry, 0, pageSize)

	log.Info().Int("page", page).Int("pageSize", pageSize).
		Msgf("[%s] start reading energy counter data", dbCounterPerfix)
	startTs := time.Now()

	rows, qErr := c.dbConn.Query(energyMeasuresQuery(), pageSize, rowsToSkip(page, pageSize))
	if qErr != nil {
		log.Error().Err(qErr).
			Int("page", page).
			Int("pageSize", pageSize).
			Msg("failed energyMeasuresQuery query")
		return energyData, qErr
	}

	var date string
	var energy float64
	for rows.Next() {
		sErr := rows.Scan(&date, &energy)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of energyMeasuresQuery", dbCounterPerfix)
			continue
		}
		energyData = append(energyData, EnergyCounterEntry{date, energy})
	}
	elapsed := time.Since(startTs)
	log.Info().Int("rowsLoaded", len(energyData)).Dur("durationMillis", elapsed).
		Msgf("[%s] finished reading energy counter data", dbCounterPerfix)

	return energyData, nil
}

func waterMeasuresQuery() string {
	return `
		SELECT
			Date,
			ColdWaterLiters,
			HotWaterLiters
		FROM
			waterCounter
		ORDER BY
			Date DESC
		LIMIT
			?
		OFFSET
			?
	`
}

func energyMeasuresQuery() string {
	return `
		SELECT
			Date,
			EnergyKwh
		FROM
			energyCounter
		ORDER BY
			Date DESC
		LIMIT
			?
		OFFSET
			?
	`
}

func rowsToSkip(page, pageSize int) int {
	if page <= 1 {
		return 0
	}
	return pageSize * (page - 1)
}
