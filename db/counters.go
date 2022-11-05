package db

type WaterCounterEntry struct {
	Date            string
	ColdWaterLiters int
	HotWaterLiters  int
}

// WaterData reads water counter entries for specified page and page size with
// descending order by the date of measurement.
func (c *Client) WaterData(page, pageSize int) ([]WaterCounterEntry, error) {
	waterData := make([]WaterCounterEntry, 0, pageSize)

	rows, qErr := c.dbConn.Query(waterMeasuresQuery(), pageSize, rowsToSkip(page, pageSize))
	if qErr != nil {
		// logger.Errorf("[%s] Query failed for technologies '[%s]': %s", TechName,
		return waterData, qErr
	}

	var date string
	var coldWater, hotWater int
	for rows.Next() {
		sErr := rows.Scan(&date, &coldWater, &hotWater)
		if sErr != nil {
			// logger.Errorf("[%s] Error while scanning query for techs '[%s]': %s",
			continue
		}
		waterData = append(waterData, WaterCounterEntry{date, coldWater, hotWater})
	}

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

	rows, qErr := c.dbConn.Query(energyMeasuresQuery(), pageSize, rowsToSkip(page, pageSize))
	if qErr != nil {
		// logger.Errorf("[%s] Query failed for technologies '[%s]': %s", TechName,
		return energyData, qErr
	}

	var date string
	var energy float64
	for rows.Next() {
		sErr := rows.Scan(&date, &energy)
		if sErr != nil {
			// logger.Errorf("[%s] Error while scanning query for techs '[%s]': %s",
			continue
		}
		energyData = append(energyData, EnergyCounterEntry{date, energy})
	}

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
