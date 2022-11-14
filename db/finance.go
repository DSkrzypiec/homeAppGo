package db

import (
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
	EndingBalanceValue    *string
	Description           string
}

// FinTransMonthly reads all financial transactions from given month.
func (c *Client) FinTransMonthly(year int, month int) ([]BankTransaction, error) {
	startTs := time.Now()
	transactions := make([]BankTransaction, 0, 500)
	log.Info().Msgf("[%s] start reading transactions for [%d-%d-*]", dbFinPrefix, year, month)

	monthStr := strconv.Itoa(month)
	if month < 10 {
		monthStr = "0" + monthStr
	}
	likeCond := fmt.Sprintf("%d-%s-%%", year, monthStr)
	log.Info().Msgf("[%s] transactions will be filtered with [LIKE '%s']", dbFinPrefix, likeCond)

	rows, qErr := c.dbConn.Query(monthlyTransactionQuery(), likeCond)
	if qErr != nil {
		log.Error().Err(qErr).Msgf("[%s] monthlyTransactionQuery failed", dbFinPrefix)
		return transactions, qErr
	}

	var id int
	var amount float64
	var accNumber, execDate, orderDate, amountCurr, description string
	var ttype, endingBalanceCurr, endingBalanceValue *string
	for rows.Next() {
		sErr := rows.Scan(&id, &accNumber, &execDate, &orderDate, &ttype, &amountCurr, &amount,
			&endingBalanceCurr, &endingBalanceValue, &description)
		if sErr != nil {
			log.Warn().Err(sErr).Msgf("[%s] while scanning results of monthlyTransactionQuery", dbFinPrefix)
			continue
		}
		transactions = append(transactions, BankTransaction{
			TransactionId:         id,
			AccountNumber:         accNumber,
			ExecutionDate:         execDate,
			OrderDate:             orderDate,
			Type:                  ttype,
			AmountCurrency:        amountCurr,
			AmountValue:           amount,
			EndingBalanceCurrency: endingBalanceCurr,
			EndingBalanceValue:    endingBalanceValue,
			Description:           description,
		})
	}
	log.Info().Int("rowsLoaded", len(transactions)).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished reading transactions for [%d-%d-*]", dbDocsPrefix, year, month)

	return transactions, nil
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
