package finance

import (
	"errors"
	"fmt"
	"homeApp/db"
	"strconv"
)

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
