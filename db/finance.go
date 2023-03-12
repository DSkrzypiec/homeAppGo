package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const dbFinPrefix = "db/fin"

type BankTransaction struct {
	TransactionId         int
	AccountNumber         string
	ExecutionDate         string
	OrderDate             string
	Type                  *string
	AmountCurrency        string
	AmountValue           float64
	EndingBalanceCurrency *string
	EndingBalanceValue    *float64
	Description           string
}

// FinTransMonthly reads all financial transactions from given month.
func (c *Client) FinTransMonthly(year int, month int) ([]BankTransaction, error) {
	log.Info().Msgf("[%s] start reading transactions for [%d-%d-*]", dbFinPrefix, year, month)

	monthStr := strconv.Itoa(month)
	if month < 10 {
		monthStr = "0" + monthStr
	}
	likeCond := fmt.Sprintf("%d-%s-%%", year, monthStr)

	return c.finQueryTransactions(monthlyTransactionQuery(), likeCond)
}

// FinInsertTransactions inserts bank transactions into database. Transactions
// is bundled into SQL transaction which is unrolled in case when there is a
// failure.
func (c *Client) FinInsertTransactions(transactions []BankTransaction) error {
	startTs := time.Now()
	log.Info().Msgf("[%s] start inserting financial transactions", dbFinPrefix)

	tx, tErr := c.dbConn.Begin()
	if tErr != nil {
		log.Error().Err(tErr).Msgf("[%s] cannot start new transaction", dbFinPrefix)
		return tErr
	}

	for _, transaction := range transactions {
		tErr := c.insertTransaction(transaction, tx)
		if tErr != nil {
			log.Error().Err(tErr).Msgf("[%s] bank transaction insertion failed", dbFinPrefix)
			rollErr := tx.Rollback()
			if rollErr != nil {
				log.Error().Err(rollErr).Msgf("[%s] SQL TX rollback failed", dbFinPrefix)
				return rollErr
			}
			return tErr
		}
	}

	commErr := tx.Commit()
	if commErr != nil {
		log.Error().Err(commErr).Msgf("[%s] couldn't commit SQL transaction", dbFinPrefix)
		rollErr := tx.Rollback()
		if rollErr != nil {
			log.Error().Err(rollErr).Msgf("[%s] SQL TX rollback failed", dbFinPrefix)
			return rollErr
		}
		return commErr
	}

	log.Info().Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished inserting financial transactions", dbFinPrefix)
	return nil
}

// FinTransByDescription reads all financial transaction that contains given
// phrase in the transaction description.
func (c *Client) FinTransByDescription(phrase string) ([]BankTransaction, error) {
	likeCond := fmt.Sprintf("%%%s%%", phrase)
	return c.finQueryTransactions(finTransactionByDescriptionQuery(), likeCond)
}

func (c *Client) finQueryTransactions(query string, args ...interface{}) ([]BankTransaction, error) {
	startTs := time.Now()
	transactions := make([]BankTransaction, 0, 500)
	log.Info().Msgf("[%s] start reading financial transactions for query [%s]", dbFinPrefix, query)

	rows, qErr := c.dbConn.Query(query, args...)
	if qErr != nil {
		log.Error().Err(qErr).Msgf("[%s] financial query failed", dbFinPrefix)
		return transactions, qErr
	}

	var id int
	var amount, endingBalanceValue float64
	var accNumber, execDate, orderDate, amountCurr, description string
	var ttype, endingBalanceCurr *string
	for rows.Next() {
		sErr := rows.Scan(&id, &accNumber, &execDate, &orderDate, &ttype, &amountCurr, &amount,
			&endingBalanceCurr, &endingBalanceValue, &description)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of [%s] query", dbFinPrefix, query)
			continue
		}
		endingBalanceCopy := endingBalanceValue
		transactions = append(transactions, BankTransaction{
			TransactionId:         id,
			AccountNumber:         accNumber,
			ExecutionDate:         execDate,
			OrderDate:             orderDate,
			Type:                  ttype,
			AmountCurrency:        amountCurr,
			AmountValue:           amount,
			EndingBalanceCurrency: endingBalanceCurr,
			EndingBalanceValue:    &endingBalanceCopy,
			Description:           description,
		})
	}
	log.Info().Int("rowsLoaded", len(transactions)).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished reading financial transactions", dbFinPrefix)

	return transactions, nil
}

// Inserts single bank transaction into database.
func (c *Client) insertTransaction(t BankTransaction, tx *sql.Tx) error {
	_, insErr := tx.Exec(
		insertTransactionQuery(), t.AccountNumber, t.ExecutionDate, t.OrderDate,
		t.Type, t.AmountCurrency, t.AmountValue, t.EndingBalanceCurrency,
		t.EndingBalanceValue, t.Description)
	return insErr
}

func monthlyTransactionQuery() string {
	return `
	SELECT
		TransactionId,
		AccountNumber,
		ExecutionDate,
		OrderDate,
		TType,
		AmountCurrency,
		AmountValue,
		EndingBalanceCurrency,
		EndingBalanceValue,
		Description
	FROM
		bankTransactions
	WHERE
		OrderDate LIKE ?
	ORDER BY
		OrderDate,
		ExecutionDate,
		AmountValue,
		AccountNumber
	`
}

func insertTransactionQuery() string {
	return `
	INSERT INTO bankTransactions (
		AccountNumber, ExecutionDate, OrderDate, TType, AmountCurrency,
		AmountValue, EndingBalanceCurrency, EndingBalanceValue, Description
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
}

func finTransactionByDescriptionQuery() string {
	return `
	SELECT
		TransactionId,
		AccountNumber,
		ExecutionDate,
		OrderDate,
		TType,
		AmountCurrency,
		AmountValue,
		EndingBalanceCurrency,
		EndingBalanceValue,
		Description
	FROM
		bankTransactions
	WHERE
		Description LIKE ?
	`
}
