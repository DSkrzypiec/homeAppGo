package finance

import (
	"errors"
	"fmt"
	"homeApp/db"
	"sort"
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

// Describe describes monthly aggregation as formatted string.
func (ma *MonthlyAgg) Describe() string {
	return fmt.Sprintf(`Transactions: %d <br>
Inflows: (%d, %.2fzł) <br>
Outflows: (%d, %.2fzł)`,
		ma.NumOfTransactions, ma.NumOfInflows, ma.InflowsAmountSum, ma.NumOfOutflows,
		ma.OutflowsAmountSum)
}

// TODO
type MonthlyChartPoint struct {
	Date           MonthDate
	MonthStr       string
	AggValueScaled float64 // needs to be from [0,1] interval
	DataLabel      string
	Tooltip        string
}

// TODO
type MonthlyChart []MonthlyChartPoint

// TODO: Docs + rename from Chart to chart or something
func AggregatesMonthlyToChart(aggs map[MonthDate]MonthlyAgg) MonthlyChart {
	chart := make(MonthlyChart, 0, len(aggs))
	maxAbsFlowVal := maxAbsMonthlyAmountSum(aggs)

	if maxAbsFlowVal == 0.0 {
		// TODO: What actually we want to do in this case?
		return chart
	}

	// TODO: Handle date filtration (to last 12 months). Here or in SQL?
	// TODO: For now we aggregate only outflows. Need to think if outflows should be selected by toggle on UI

	for monthDate, agg := range aggs {
		newPoint := MonthlyChartPoint{
			Date:           monthDate,
			MonthStr:       monthDate.String(),
			AggValueScaled: (-1.0 * agg.OutflowsAmountSum) / maxAbsFlowVal,
			DataLabel:      fmt.Sprintf("%.2fzł", agg.OutflowsAmountSum),
			Tooltip:        agg.Describe(),
		}
		chart = append(chart, newPoint)
	}

	sort.SliceStable(chart, func(i, j int) bool {
		if chart[i].Date.Year == chart[j].Date.Year {
			return chart[i].Date.Month < chart[j].Date.Month
		}
		return chart[i].Date.Year < chart[j].Date.Year
	})

	return chart
}

// Calculates max absolute value based on both InflowsAmountSum and OutflowsAmountSum across all monthly aggregations.
func maxAbsMonthlyAmountSum(aggs map[MonthDate]MonthlyAgg) float64 {
	maxAbs := 0.0
	for _, agg := range aggs {
		// TODO: For now we aggregate only outflows. Need to think if outflows should be selected by toggle on UI
		/*
			if agg.InflowsAmountSum > maxAbs {
				maxAbs = agg.InflowsAmountSum
			}
		*/
		if -1.0*agg.OutflowsAmountSum > maxAbs {
			maxAbs = -1.0 * agg.OutflowsAmountSum
		}
	}
	return maxAbs
}

// AggregateMonthly performs monthly grouping on given set of BankTransactions.
func AggregateMonthly(transactions []db.BankTransaction, defaultCurrency string) map[MonthDate]MonthlyAgg {
	aggs := make(map[MonthDate]MonthlyAgg)
	monthlyGroups := groupTransMonthly(transactions)
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
			// TODO: This will be handeled in the future. Probably by performing additional aggregation by currency.
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
func groupTransMonthly(trans []db.BankTransaction) map[MonthDate][]db.BankTransaction {
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
