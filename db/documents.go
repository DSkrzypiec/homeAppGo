package db

import (
	"time"

	"github.com/rs/zerolog/log"
)

const dbDocsPrefix = "db/documents"

type DocumentInfo struct {
	Id             int
	Name           string
	UploadDate     string
	DocumentDate   string
	Category       string
	PersonInvolved string
	FileExtension  string
	FileSizeBytes  int64
}

// Documents reads list of all metadata of documents (without content itself).
func (c *Client) Documents() ([]DocumentInfo, error) {
	startTs := time.Now()
	documents := make([]DocumentInfo, 0, 500)
	log.Info().Msgf("[%s] start reading water counter data", dbDocsPrefix)

	rows, qErr := c.dbConn.Query(documentsQuery())
	if qErr != nil {
		log.Error().Err(qErr).Msgf("[%s] documentsQuery failed", dbDocsPrefix)
		return documents, qErr
	}

	var id int
	var fileSize int64
	var name, uploadDate, documentDate, category, person, fileExt string
	for rows.Next() {
		sErr := rows.Scan(&id, &name, &uploadDate, &documentDate, &category, &person, &fileExt, &fileSize)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of documentsQuery", dbDocsPrefix)
			continue
		}
		documents = append(documents, DocumentInfo{
			Id:             id,
			Name:           name,
			UploadDate:     uploadDate,
			DocumentDate:   documentDate,
			Category:       category,
			PersonInvolved: person,
			FileExtension:  fileExt,
			FileSizeBytes:  fileSize,
		})
	}
	log.Info().Int("rowsLoaded", len(documents)).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished reading documents info", dbDocsPrefix)

	return documents, nil
}

func documentsQuery() string {
	return `
		SELECT
			DocumentId,
			DocumentName,
			UploadDate,
			DocumentDate,
			Category,
			PersonInvolved,
			FileExtension,
			FileSize
		FROM
			documents
		ORDER BY
			DocumentId DESC
	`
}
