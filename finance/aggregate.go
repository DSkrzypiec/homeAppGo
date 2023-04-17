package finance

import (
	"errors"
	"fmt"
	"homeApp/db"
	"strconv"
)

// MonthlyAgg represents statistics regarding monthly batch of financial transactions.
type MonthlyAgg struct {
	MonthDate         string
	NumOfTransactions int
	NumOfInflows      int
	NumOfOutflows     int
	InflowsAmountSum  float64
	OutflowsAmountSum float64
	// TODO: Add aggregates for other currencies

	// "Top" here means by absolute value of AmountValue
	TopInflow  db.BankTransaction
	TopOutflow db.BankTransaction
}

// AggregateMonthly performs monthly grouping on given set of BankTransactions.
func AggregateMonthly(transactions []db.BankTransaction, defaultCurrency string) map[MonthDate]MonthlyAgg {
	aggs := make(map[MonthDate]MonthlyAgg)
	monthlyGroups := GroupTransMonthly(transactions)
	for monthDate, trans := range monthlyGroups {
		aggs[monthDate] = aggregateSingleMonth(monthDate.String(), defaultCurrency, trans)
	}
	return aggs
}

// Aggregates transactions from single month into MonthlyAgg.
func aggregateSingleMonth(dateMonth string, defaultCurrency string, monthTransactions []db.BankTransaction) MonthlyAgg {
	var (
		inflows, outflows             int
		inflowsAmount, outflowsAmount float64
		topInflowTransaction          db.BankTransaction
		topOutflowTransaction         db.BankTransaction
	)
	for _, t := range monthTransactions {
		if t.AmountCurrency != defaultCurrency {
			// TODO: This will be handeled in the future
			continue
		}

		if t.AmountValue >= 0 {
			inflows++
			inflowsAmount += t.AmountValue
			if t.AmountValue > topInflowTransaction.AmountValue {
				tmp := t
				topInflowTransaction = tmp
			}
		} else {
			outflows++
			outflowsAmount += t.AmountValue
			if t.AmountValue < topOutflowTransaction.AmountValue {
				tmp := t
				topOutflowTransaction = tmp
			}
		}
	}

	return MonthlyAgg{
		MonthDate:         dateMonth,
		NumOfTransactions: len(monthTransactions),
		NumOfInflows:      inflows,
		NumOfOutflows:     outflows,
		InflowsAmountSum:  inflowsAmount,
		OutflowsAmountSum: outflowsAmount,
		TopInflow:         topInflowTransaction,
		TopOutflow:        topOutflowTransaction,
	}
}

// GroupTransMonthly groups transactions in monthly slices. Grouping is done by
// OrderDate field.
func GroupTransMonthly(trans []db.BankTransaction) map[MonthDate][]db.BankTransaction {
	dateToTrans := make(map[MonthDate][]db.BankTransaction)

	for _, t := range trans {
		date, dErr := parseMonthDate(t.OrderDate)
		if dErr != nil {
			// Correct date format in BankTransaction is assumed. Data with
			// incorrect date format will be ignored in the aggregation.
			continue
		}
		if _, dateExist := dateToTrans[date]; !dateExist {
			dateToTrans[date] = make([]db.BankTransaction, 0, 100)
		}
		dateToTrans[date] = append(dateToTrans[date], t)
	}

	return dateToTrans
}

// YYYY-MM date. For now usecase is too simple to use package time.
type MonthDate struct {
	Year  int
	Month int
}

// String representation.
func (md *MonthDate) String() string {
	if md.Month < 10 {
		return fmt.Sprintf("%d-0%d", md.Year, md.Month)
	}
	return fmt.Sprintf("%d-%d", md.Year, md.Month)
}

// Parses MonthDate based on string. Assuming YYYY-MM-dd format.
func parseMonthDate(input string) (MonthDate, error) {
	if len(input) < 7 {
		return MonthDate{}, errors.New("given input is too short, expected at least 7 chars")
	}
	if input[4] != '-' {
		return MonthDate{}, errors.New("incorrect input form, expected 'YYYY-MM'")
	}
	year, yearErr := strconv.Atoi(input[0:4])
	if yearErr != nil {
		return MonthDate{}, fmt.Errorf("cannot parse year as int: %w", yearErr)
	}
	if year < 1900 {
		return MonthDate{}, fmt.Errorf("expected year >= 1900, got: %d", year)
	}
	month, monthErr := strconv.Atoi(input[5:7])
	if monthErr != nil {
		return MonthDate{}, fmt.Errorf("cannot parse month as int: %w", monthErr)
	}
	if month < 1 || month > 12 {
		return MonthDate{}, fmt.Errorf("incorrect month value, expected [1, 12], got: %d", month)
	}
	return MonthDate{Year: year, Month: month}, nil
}
