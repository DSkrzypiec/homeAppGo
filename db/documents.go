package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

const dbDocsPrefix = "db/documents"

// DocumentInfo represents document file metadata.
type DocumentInfo struct {
	Id             int
	Name           string
	UploadDate     string
	DocumentDate   *string
	Category       string
	PersonInvolved *string
	FileExtension  string
	FileSizeBytes  int64
}

// NewDocument represents new documentat candidate that shall be inserted into
// Home DB.
type NewDocument struct {
	Name           string
	DocumentDate   string
	Category       string
	PersonInvolved string
	FileExtension  string
	DocumentFile   []byte
}

// Documents reads list of all metadata of documents (without content itself).
func (c *Client) Documents() ([]DocumentInfo, error) {
	startTs := time.Now()
	documents := make([]DocumentInfo, 0, 500)
	log.Info().Msgf("[%s] start reading documents metadata", dbDocsPrefix)

	rows, qErr := c.dbConn.Query(documentsQuery())
	if qErr != nil {
		log.Error().Err(qErr).Msgf("[%s] documentsQuery failed", dbDocsPrefix)
		return documents, qErr
	}

	var id int
	var fileSize int64
	var name, uploadDate, category, fileExt string
	var documentDate, person *string
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

// DocumentsFiltered reads filtered list of all metadata of documents (without content itself).
// Filtering is done by full text search in the DB using FTS 5.
func (c *Client) DocumentsFiltered(phrase string) ([]DocumentInfo, error) {
	startTs := time.Now()
	documents := make([]DocumentInfo, 0, 500)
	log.Info().Str("filter", phrase).Msgf("[%s] start reading filtered documents metadata", dbDocsPrefix)

	queryPhrase := fmt.Sprintf(`"%s"*`, phrase)
	rows, qErr := c.dbConn.Query(documentsFilteredQuery(), queryPhrase)
	if qErr != nil {
		log.Error().Err(qErr).Msgf("[%s] documentsFilteredQuery failed", dbDocsPrefix)
		return documents, qErr
	}

	var id int
	var fileSize int64
	var name, uploadDate, category, fileExt string
	var documentDate, person *string
	for rows.Next() {
		sErr := rows.Scan(&id, &name, &uploadDate, &documentDate, &category, &person, &fileExt, &fileSize)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of documentsFilteredQuery", dbDocsPrefix)
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
	log.Info().Str("filter", phrase).Int("rowsLoaded", len(documents)).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished reading filtered documents info", dbDocsPrefix)

	return documents, nil
}

// DocumentInsertNew uploads into database information about new document. This
// is composed of two things - metadata about document and the content.
func (c *Client) DocumentInsertNew(newDoc NewDocument) error {
	startTs := time.Now()
	log.Info().Msgf("[%s] start inserting new document", dbDocsPrefix)

	tx, tErr := c.dbConn.Begin()
	if tErr != nil {
		log.Error().Err(tErr).Msgf("[%s] cannot start new transaction", dbDocsPrefix)
		return tErr
	}

	maxDocId, diErr := getMaxDocumentId(tx)
	if diErr != nil {
		log.Error().Err(diErr).Msgf("[%s] cannot get max DocumentId", dbDocsPrefix)
		return diErr
	}
	newDocId := maxDocId + 1

	metaErr := insertDocumentMeta(newDocId, newDoc, tx)
	if metaErr != nil {
		log.Error().Err(metaErr).Msgf("[%s] cannot insert document metadata", dbDocsPrefix)
		return metaErr
	}

	fts5Err := insertDocumentFts5(newDocId, newDoc, tx)
	if fts5Err != nil {
		log.Error().Err(fts5Err).Msgf("[%s] cannot insert document FTS5 info", dbDocsPrefix)
		return fts5Err
	}

	contentErr := insertDocumentContent(newDocId, newDoc.DocumentFile, tx)
	if contentErr != nil {
		log.Error().Err(contentErr).Msgf("[%s] cannot insert document content", dbDocsPrefix)
		return contentErr
	}

	commErr := tx.Commit()
	if commErr != nil {
		log.Error().Err(commErr).Msgf("[%s] could not commit transaction, rollback.", dbDocsPrefix)
		tx.Rollback()
		return commErr
	}

	log.Info().Dur("duration", time.Since(startTs)).Int64("fileSize", int64(len(newDoc.DocumentFile))).
		Msgf("[%s] finished inserting new document", dbDocsPrefix)

	return nil
}

// DocumentFile load from database content of given document and it's extension
// name.
func (c *Client) DocumentFile(documentId int) ([]byte, string, error) {
	startTs := time.Now()
	log.Info().Int("documentId", documentId).Msgf("[%s] start reading document file", dbDocsPrefix)

	row := c.dbConn.QueryRow(documentFileQuery(), documentId)
	var fileBytes []byte
	var fileExt string
	scanErr := row.Scan(&fileExt, &fileBytes)
	if scanErr != nil {
		log.Error().Err(scanErr).Int("documentId", documentId).
			Msgf("[%s] while loading document content", dbDocsPrefix)
		return fileBytes, fileExt, scanErr
	}

	log.Info().Int("loadedBytes", len(fileBytes)).Int("documentId", documentId).
		Dur("duration", time.Since(startTs)).Msgf("[%s] finished loading document content", dbDocsPrefix)

	return fileBytes, fileExt, nil
}

func getMaxDocumentId(tx *sql.Tx) (int, error) {
	var maxDocId int
	row := tx.QueryRow("SELECT CASE WHEN MAX(DocumentId) IS NULL THEN 0 ELSE MAX(DocumentId) END FROM documents")
	scanErr := row.Scan(&maxDocId)
	if scanErr != nil {
		return -1, scanErr
	}
	return maxDocId, nil
}

func insertDocumentMeta(documentId int, newDoc NewDocument, tx *sql.Tx) error {
	validErr := newDoc.Validate()
	if validErr != nil {
		return validErr
	}

	uploadDate := time.Now().Format("2006-01-02")

	_, qErr := tx.Exec(
		documentInsertNewMetaQuery(), documentId, newDoc.Name, uploadDate,
		toNullString(&newDoc.DocumentDate), newDoc.Category, toNullString(&newDoc.PersonInvolved),
		newDoc.FileExtension, len(newDoc.DocumentFile))
	if qErr != nil {
		return qErr
	}
	return nil
}

func insertDocumentFts5(documentId int, newDoc NewDocument, tx *sql.Tx) error {
	uploadDate := time.Now().Format("2006-01-02")
	_, qErr := tx.Exec(
		documentInsertNewFts5Query(), documentId, newDoc.Name, uploadDate,
		toNullString(&newDoc.DocumentDate), newDoc.Category,
		toNullString(&newDoc.PersonInvolved), newDoc.FileExtension)
	if qErr != nil {
		return qErr
	}
	return nil
}

func insertDocumentContent(documentId int, content []byte, tx *sql.Tx) error {
	_, qErr := tx.Exec(documentInsertNewQuery(), documentId, content)
	if qErr != nil {
		return qErr
	}
	return nil
}

func (nd *NewDocument) Validate() error {
	if len(nd.Name) == 0 {
		return errors.New("document name should be provided")
	}
	if len(nd.Category) == 0 {
		return errors.New("document category should be provided")
	}
	if len(nd.FileExtension) == 0 {
		return errors.New("document extension should be provided")
	}
	return nil
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

func documentsFilteredQuery() string {
	return `
	SELECT
		d.DocumentId,
		d.DocumentName,
		d.UploadDate,
		d.DocumentDate,
		d.Category,
		d.PersonInvolved,
		d.FileExtension,
		d.FileSize
	FROM
		documentsFts5 f
	INNER JOIN
		documents d ON f.DocumentId = d.DocumentId
	WHERE
		f.documentsFts5 MATCH ?
	`
}

func documentInsertNewMetaQuery() string {
	return `
	INSERT INTO documents (
		DocumentId, DocumentName, UploadDate, DocumentDate, Category, PersonInvolved,
		FileExtension, FileSize
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
}

func documentInsertNewFts5Query() string {
	return `
	INSERT INTO documentsFts5 (
		DocumentId, DocumentName, UploadDate, DocumentDate, Category,
		PersonInvolved, FileExtension
	)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
}

func documentInsertNewQuery() string {
	return `
		INSERT INTO documentFiles (DocumentId, FileBytes)
		VALUES (?, ?)
	`
}

func documentFileQuery() string {
	return `
		SELECT
			d.FileExtension,
			df.FileBytes
		FROM
			documents d
		INNER JOIN
			documentFiles df ON d.DocumentId = df.DocumentId
		WHERE
			d.DocumentId = ?
`
}
