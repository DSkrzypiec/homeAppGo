package db

import (
	"time"

	"github.com/rs/zerolog/log"
)

const dbHomePrefix = "db/home"

type HomeSummary struct {
	ColdWaterWeeklyAvg       float64
	HotWaterWeeklyAvg        float64
	EnergyWeeklyAvg          float64
	DocumentsNumber          int
	DocumentsSizeMb          float64
	DocumentLatestUploadDate string
	EbooksNumber             int
	EbooksSizeMb             float64
	EbookLatestUploadDate    string
	BooksNumber              int
	BookLatestUploadDate     string
	FinancialLatestOrderDate string
}

// HomeSummary calculates and reads home database summary.
func (c *Client) HomeSummary() (HomeSummary, error) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start reading home summary", dbHomePrefix)

	row := c.dbConn.QueryRow(homeSummaryQuery())

	var docNum, ebookNum, bookNum int
	var docLatestDate, ebookLatestDate, bookLatestDate, finLatestDate string
	var coldWaterAvg, hotWaterAvg, energyAvg, docSizeMb, ebookSizeMb float64

	sErr := row.Scan(
		&coldWaterAvg, &hotWaterAvg, &energyAvg,
		&docNum, &docSizeMb, &docLatestDate,
		&ebookNum, &ebookSizeMb, &ebookLatestDate,
		&bookNum, &bookLatestDate,
		&finLatestDate)
	if sErr != nil {
		log.Error().Err(sErr).Msgf("[%s] cannot scan homeSummaryQuery result", dbHomePrefix)
		return HomeSummary{}, sErr
	}

	summary := HomeSummary{
		ColdWaterWeeklyAvg:       coldWaterAvg,
		HotWaterWeeklyAvg:        hotWaterAvg,
		EnergyWeeklyAvg:          energyAvg,
		DocumentsNumber:          docNum,
		DocumentsSizeMb:          docSizeMb,
		DocumentLatestUploadDate: docLatestDate,
		EbooksNumber:             ebookNum,
		EbooksSizeMb:             ebookSizeMb,
		EbookLatestUploadDate:    ebookLatestDate,
		BooksNumber:              bookNum,
		BookLatestUploadDate:     bookLatestDate,
		FinancialLatestOrderDate: finLatestDate,
	}

	log.Info().Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished reading home summary", dbHomePrefix)

	return summary, nil
}

func homeSummaryQuery() string {
	return `
	WITH countersRowId AS (
		SELECT
			ROW_NUMBER() OVER (ORDER BY wc.Date DESC) AS RowId,
			wc.Date,
			wc.ColdWaterLiters,
			wc.HotWaterLiters,
			ec.EnergyKwh
		FROM
			waterCounter wc
		INNER JOIN
			energyCounter ec ON wc.Date = ec.Date
	)
	, countersLatest AS (
		SELECT Date, ColdWaterLiters, HotWaterLiters, EnergyKwh FROM countersRowId WHERE RowId = 1
	)
	, countersPrev AS (
		SELECT Date, ColdWaterLiters, HotWaterLiters, EnergyKwh FROM countersRowId WHERE RowId = 3
	)
	, countersSummary AS (
		SELECT
			(latest.ColdWaterLiters - COALESCE(prev.ColdWaterLiters, 0)) /
			COALESCE(JULIANDAY(latest.Date) - JULIANDAY(prev.Date) + 1, 1.0) * 7.0 AS ColdWaterWeeklyAvg,
			(latest.HotWaterLiters - COALESCE(prev.HotWaterLiters, 0)) /
			COALESCE(JULIANDAY(latest.Date) - JULIANDAY(prev.Date) + 1, 1.0) * 7.0 AS HotWaterWeeklyAvg,
			(latest.EnergyKwh - COALESCE(prev.EnergyKwh, 0)) /
			COALESCE(JULIANDAY(latest.Date) - JULIANDAY(prev.Date) + 1, 1.0) * 7.0 AS EnergyWeeklyAvg
		FROM
			countersLatest latest
		CROSS JOIN
			countersPrev prev
	)
	, documentsSummary AS (
		SELECT
			COUNT(d.DocumentId) AS DocNumber,
			SUM(LENGTH(df.FileBytes)) / 1000000.0 AS DocSizeMb,
			MAX(d.UploadDate) AS DocLatestDate
		FROM
			documents d
		INNER JOIN
			documentFiles df ON d.DocumentId = df.DocumentId
	)
	, booksSummary AS (
		SELECT
			SUM(
				CASE
					WHEN FileSize IS NOT NULL THEN 1
					ELSE 0
				END
			) AS EbooksNumber,
			CASE
				WHEN SUM(FileSize) IS NULL THEN 0.0
				ELSE SUM(FileSize) / 1000000.0
			END AS EbooksSizeMb,
			MAX(
				CASE
					WHEN FileSize IS NOT NULL THEN UploadDate
					ELSE ''
				END
			) AS EbookLatestUploadDate,
			SUM(
				CASE
					WHEN FileSize IS NULL THEN 1
					ELSE 0
				END
			) AS BooksNumber,
			MAX(
				CASE
					WHEN FileSize IS NULL THEN UploadDate
					ELSE ''
				END
			) AS BookLatestUploadDate
		FROM
			books b
	)
	, finSummary AS (
		SELECT
			MAX(OrderDate) AS FinLatestOrderDate
		FROM
			bankTransactions
		WHERE
			DATE(OrderDate) IS NOT NULL  -- only ISO date format
	)
	SELECT
		cs.ColdWaterWeeklyAvg,
		cs.HotWaterWeeklyAvg,
		cs.EnergyWeeklyAvg,
		ds.DocNumber,
		ds.DocSizeMb,
		ds.DocLatestDate,
		bs.EbooksNumber,
		bs.EbooksSizeMb,
		bs.EbookLatestUploadDate,
		bs.BooksNumber,
		bs.BookLatestUploadDate,
		fs.FinLatestOrderDate
	FROM
		countersSummary cs
	CROSS JOIN
		documentsSummary ds
	CROSS JOIN
		booksSummary bs
	CROSS JOIN
		finSummary fs
	;
	`
}
