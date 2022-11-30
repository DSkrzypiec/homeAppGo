package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

const dbBooksPrefix = "db/books"

type Book struct {
	Id             int
	Title          string
	Authors        string
	Publisher      *string
	PublishingYear *int
	Category       string
	Language       string
	FileExtension  *string
	FileSize       *int
	UploadDate     string
}

// NewBook represents new book candidate that shall be inserted into Home DB.
type NewBook struct {
	Title          string
	Authors        string
	Publisher      *string
	PublishingYear *int
	Category       string
	Language       string
	FileExtension  *string
	FileSize       *int
	BookFile       []byte
}

// Books reads list of all metadata of documents (without content itself).
func (c *Client) Books() ([]Book, error) {
	startTs := time.Now()
	books := make([]Book, 0, 500)
	log.Info().Msgf("[%s] start reading books metadata", dbBooksPrefix)

	rows, qErr := c.dbConn.Query(booksQuery())
	if qErr != nil {
		log.Error().Err(qErr).Msgf("[%s] booksQuery failed", dbBooksPrefix)
		return books, qErr
	}

	var id int
	var publishingYear, fileSize *int
	var title, authors, category, language, uploadDate string
	var publisher, fileExt *string
	for rows.Next() {
		sErr := rows.Scan(&id, &title, &authors, &publisher, &publishingYear, &category, &language,
			&fileExt, &fileSize, &uploadDate)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of booksQuery", dbBooksPrefix)
			continue
		}
		books = append(books, Book{
			Id:             id,
			Title:          title,
			Authors:        authors,
			Publisher:      publisher,
			PublishingYear: publishingYear,
			Category:       category,
			Language:       language,
			FileExtension:  fileExt,
			FileSize:       fileSize,
			UploadDate:     uploadDate,
		})
	}
	log.Info().Int("rowsLoaded", len(books)).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished reading books info", dbBooksPrefix)

	return books, nil
}

// Books reads list of filtered metadata of documents (without content itself)
// by given phrase.
func (c *Client) BooksFiltered(phrase string) ([]Book, error) {
	startTs := time.Now()
	books := make([]Book, 0, 500)
	log.Info().Str("filter", phrase).Msgf("[%s] start reading filtered books metadata", dbBooksPrefix)

	queryPhrase := fmt.Sprintf(`"%s"*`, phrase)
	rows, qErr := c.dbConn.Query(booksFilteredQuery(), queryPhrase)
	if qErr != nil {
		log.Error().Err(qErr).Str("filter", phrase).Msgf("[%s] booksQuery failed", dbBooksPrefix)
		return books, qErr
	}

	var id int
	var publishingYear, fileSize *int
	var title, authors, category, language, uploadDate string
	var publisher, fileExt *string
	for rows.Next() {
		sErr := rows.Scan(&id, &title, &authors, &publisher, &publishingYear, &category,
			&language, &fileExt, &fileSize, &uploadDate)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of booksQuery", dbBooksPrefix)
			continue
		}
		books = append(books, Book{
			Id:             id,
			Title:          title,
			Authors:        authors,
			Publisher:      publisher,
			PublishingYear: publishingYear,
			Category:       category,
			Language:       language,
			FileExtension:  fileExt,
			FileSize:       fileSize,
			UploadDate:     uploadDate,
		})
	}
	log.Info().Int("rowsLoaded", len(books)).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished reading books info", dbBooksPrefix)

	return books, nil
}

// BookInsertNew uploads into database information about new book. This
// is composed of two things - metadata about book and possibly (if e-book with
// provided file) the content.
func (c *Client) BookInsertNew(newBook NewBook) error {
	startTs := time.Now()
	log.Info().Msgf("[%s] start inserting new book", dbBooksPrefix)

	tx, tErr := c.dbConn.Begin()
	if tErr != nil {
		log.Error().Err(tErr).Msgf("[%s] cannot start new transaction", dbBooksPrefix)
		return tErr
	}

	maxBookId, diErr := getMaxBookId(tx)
	if diErr != nil {
		log.Error().Err(diErr).Msgf("[%s] cannot get max BookId", dbBooksPrefix)
		return diErr
	}
	newBookId := maxBookId + 1

	metaErr := insertBookMeta(newBookId, newBook, tx)
	if metaErr != nil {
		log.Error().Err(metaErr).Msgf("[%s] cannot insert book metadata", dbBooksPrefix)
		return metaErr
	}

	fts5Err := insertBookFts5(newBookId, newBook, tx)
	if fts5Err != nil {
		log.Error().Err(fts5Err).Msgf("[%s] cannot insert book FTS5 info", dbBooksPrefix)
		return fts5Err
	}

	if len(newBook.BookFile) > 0 {
		contentErr := insertBookContent(newBookId, newBook.BookFile, tx)
		if contentErr != nil {
			log.Error().Err(contentErr).Msgf("[%s] cannot insert book content", dbBooksPrefix)
			return contentErr
		}
	}

	commErr := tx.Commit()
	if commErr != nil {
		log.Error().Err(commErr).Msgf("[%s] could not commit transaction, rollback.", dbBooksPrefix)
		tx.Rollback()
		return commErr
	}

	log.Info().Dur("duration", time.Since(startTs)).Int64("fileSize", int64(len(newBook.BookFile))).
		Msgf("[%s] finished inserting new document", dbBooksPrefix)

	return nil
}

// BookFile load from database content of given book, it's title and it's extension
// name.
func (c *Client) BookFile(bookId int) ([]byte, string, string, error) {
	startTs := time.Now()
	log.Info().Int("bookId", bookId).Msgf("[%s] start reading book file", dbBooksPrefix)

	row := c.dbConn.QueryRow(bookFileQuery(), bookId)
	var fileBytes []byte
	var fileExt, title string
	scanErr := row.Scan(&fileExt, &title, &fileBytes)
	if scanErr != nil {
		log.Error().Err(scanErr).Int("bookId", bookId).
			Msgf("[%s] while loading book content", dbDocsPrefix)
		return fileBytes, title, fileExt, scanErr
	}

	log.Info().Int("loadedBytes", len(fileBytes)).Int("bookId", bookId).
		Dur("duration", time.Since(startTs)).Msgf("[%s] finished loading book content", dbDocsPrefix)

	return fileBytes, title, fileExt, nil
}

func getMaxBookId(tx *sql.Tx) (int, error) {
	var maxBookId int
	row := tx.QueryRow("SELECT CASE WHEN MAX(BookId) IS NULL THEN 0 ELSE MAX(BookId) END FROM books")
	scanErr := row.Scan(&maxBookId)
	if scanErr != nil {
		return -1, scanErr
	}
	return maxBookId, nil
}

func insertBookMeta(bookId int, newBook NewBook, tx *sql.Tx) error {
	validErr := newBook.Validate()
	if validErr != nil {
		return validErr
	}

	uploadDate := time.Now().Format("2006-01-02")

	_, qErr := tx.Exec(
		bookInsertNewMetaQuery(), bookId, newBook.Title, newBook.Authors,
		toNullString(newBook.Publisher), toNullInt(newBook.PublishingYear), newBook.Category,
		newBook.Language, toNullString(newBook.FileExtension), toNullInt(newBook.FileSize), uploadDate)
	if qErr != nil {
		return qErr
	}
	return nil
}

func insertBookFts5(bookId int, newBook NewBook, tx *sql.Tx) error {
	uploadDate := time.Now().Format("2006-01-02")
	_, qErr := tx.Exec(
		bookInsertNewFts5Query(), bookId, newBook.Title, newBook.Authors,
		toNullString(newBook.Publisher), toNullInt(newBook.PublishingYear),
		newBook.Category, newBook.Language, toNullString(newBook.FileExtension), uploadDate)
	if qErr != nil {
		return qErr
	}
	return nil
}

func insertBookContent(bookId int, content []byte, tx *sql.Tx) error {
	_, qErr := tx.Exec(bookInsertNewQuery(), bookId, content)
	if qErr != nil {
		return qErr
	}
	return nil
}

func (nb *NewBook) Validate() error {
	if len(nb.Title) == 0 {
		return errors.New("book title should be provided")
	}
	if len(nb.Authors) == 0 {
		return errors.New("book authors should be provided")
	}
	if len(nb.Category) == 0 {
		return errors.New("book category should be provided")
	}
	if len(nb.Language) == 0 {
		return errors.New("book language should be provided")
	}
	if len(nb.BookFile) > 0 {
		if nb.FileExtension == nil || len(*nb.FileExtension) == 0 {
			return errors.New("file extension should be provided")
		}
	}
	return nil
}

func booksQuery() string {
	return `
		SELECT
			BookId,
			Title,
			Authors,
			Publisher,
			PublishingYear,
			Category,
			Language,
			FileExtension,
			FileSize,
			UploadDate
		FROM
			books
		ORDER BY
			BookId DESC
	`
}

func booksFilteredQuery() string {
	return `
		SELECT
			b.BookId,
			b.Title,
			b.Authors,
			b.Publisher,
			b.PublishingYear,
			b.Category,
			b.Language,
			b.FileExtension,
			b.FileSize,
			b.UploadDate
		FROM
			booksFts5 f
		INNER JOIN
			books b ON f.BookId = b.BookId
		WHERE
			f.booksFts5 MATCH ?
		ORDER BY
			b.BookId DESC
	`
}

func bookInsertNewMetaQuery() string {
	return `
	INSERT INTO books (
		BookId, Title, Authors, Publisher, PublishingYear, Category, Language,
		FileExtension, FileSize, UploadDate
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
}

func bookInsertNewFts5Query() string {
	return `
	INSERT INTO booksFts5 (
		BookId, Title, Authors, Publisher, PublishingYear, Category,
		Language, FileExtension, UploadDate
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
}

func bookInsertNewQuery() string {
	return `
		INSERT INTO bookFiles (bookId, FileBytes)
		VALUES (?, ?)
	`
}

func bookFileQuery() string {
	return `
		SELECT
			b.FileExtension,
			b.Title,
			bf.FileBytes
		FROM
			books b
		INNER JOIN
			bookFiles bf ON b.BookId = bf.BookId
		WHERE
			b.BookId = ?
	`
}
