package finance

import (
	"homeApp/db"
	"testing"
)

func TestAggregateSingleMonth(t *testing.T) {
	// OrderDate is omitted, because it's assumed that those transactions are from the same month.
	ts := []db.BankTransaction{
		{TransactionId: 1, AmountCurrency: "PLN", AmountValue: -100.0},
		{TransactionId: 2, AmountCurrency: "PLN", AmountValue: -200.0},
		{TransactionId: 3, AmountCurrency: "PLN", AmountValue: 2000.0},
		{TransactionId: 4, AmountCurrency: "USD", AmountValue: -20.0},
		{TransactionId: 5, AmountCurrency: "EUR", AmountValue: 10.0},
		{TransactionId: 6, AmountCurrency: "PLN", AmountValue: 1000.0},
		{TransactionId: 7, AmountCurrency: "PLN", AmountValue: -100.0},
		{TransactionId: 8, AmountCurrency: "PLN", AmountValue: -50.0},
		{TransactionId: 9, AmountCurrency: "PLN", AmountValue: 0.0},
	}

	dateStr := "2023-01"
	defaultCurrency := "PLN"
	stats := aggregateSingleMonth(dateStr, defaultCurrency, ts)

	if stats.MonthDate != dateStr {
		t.Errorf("Expected MonthDate=%s, but got %s", dateStr, stats.MonthDate)
	}
	if stats.NumOfTransactions != 9 {
		t.Errorf("Expected 9 transactions, got: %d", stats.NumOfTransactions)
	}
	if stats.NumOfInflows != 3 {
		t.Errorf("Expected 3 inflows, got: %d", stats.NumOfInflows)
	}
	if stats.NumOfOutflows != 4 {
		t.Errorf("Expected 4 outflows, got: %d", stats.NumOfOutflows)
	}
	if stats.InflowsAmountSum != 3000.0 {
		t.Errorf("Expected inflows sums up to 3000.0, but got: %f", stats.InflowsAmountSum)
	}
	if stats.OutflowsAmountSum != -450.0 {
		t.Errorf("Expected outflows sums up to -450.0, but got: %f", stats.OutflowsAmountSum)
	}
	if stats.TopInflow != ts[2] {
		t.Errorf("Expected top inflow to be %v, but got: %v", ts[2], stats.TopInflow)
	}
	if stats.TopOutflow != ts[1] {
		t.Errorf("Expected top outflow to be %v, but got: %v", ts[1], stats.TopOutflow)
	}
}

func TestMaxAbsMonthlyAmountSum(t *testing.T) {
	ts := []db.BankTransaction{
		{TransactionId: 1, AmountCurrency: "PLN", OrderDate: "2023-01-01", AmountValue: -100.0},
		{TransactionId: 2, AmountCurrency: "PLN", OrderDate: "2023-01-21", AmountValue: -200.0},
		{TransactionId: 3, AmountCurrency: "PLN", OrderDate: "2023-02-01", AmountValue: 2000.0},
		{TransactionId: 4, AmountCurrency: "USD", OrderDate: "2023-02-09", AmountValue: -20.0},
		{TransactionId: 5, AmountCurrency: "EUR", OrderDate: "2023-03-01", AmountValue: 10.0},
		{TransactionId: 6, AmountCurrency: "PLN", OrderDate: "2023-03-10", AmountValue: 1000.0},
		{TransactionId: 7, AmountCurrency: "PLN", OrderDate: "2023-01-01", AmountValue: -100.0},
		{TransactionId: 8, AmountCurrency: "PLN", OrderDate: "2023-04-01", AmountValue: -50.0},
		{TransactionId: 9, AmountCurrency: "PLN", OrderDate: "2023-04-05", AmountValue: 0.0},
	}

	aggs := AggregateMonthly(ts, "PLN")
	maxAbs := maxAbsMonthlyAmountSum(aggs)

	if maxAbs != 400.0 {
		t.Errorf("Expected absolute max of flows to be %f, got %f", 400.0, maxAbs)
	}
}

func TestGroupTransMonthlyBasic(t *testing.T) {
	ts := []db.BankTransaction{
		{TransactionId: 1, OrderDate: "2023-01-15"},
		{TransactionId: 2, OrderDate: "2023-01-31"},
		{TransactionId: 3, OrderDate: "2022-12-01"},
		{TransactionId: 4, OrderDate: "2022-11-01"},
		{TransactionId: 5, OrderDate: "2022/11/02"}, // this should be skipped
		{TransactionId: 6, OrderDate: "2023-01-01"},
	}

	grouped := groupTransMonthly(ts)
	if len(grouped) != 3 {
		t.Errorf("Expected grouped transaction to have 3 months, but got: %d", len(grouped))
	}

	jan, janExist := grouped[MonthDate{Year: 2023, Month: 1}]
	if !janExist {
		t.Error("Expected existance of 2023.01 in grouped transaction")
	}
	if len(jan) != 3 {
		t.Errorf("Expected 3 transaction in 2023.01, but got: %d (%v)", len(jan), jan)
	}
	if jan[0].TransactionId != 1 {
		t.Errorf("Expected first transaction of 2023.01 to have TransactionId=1, but got: %d",
			jan[0].TransactionId)
	}
	if jan[2].TransactionId != 6 {
		t.Errorf("Expected third transaction of 2023.01 to have TransactionId=6, but got: %d",
			jan[2].TransactionId)
	}

	dec, decExist := grouped[MonthDate{Year: 2022, Month: 12}]
	if !decExist {
		t.Error("Expected existance of 2022.12 in grouped transaction")
	}
	if len(dec) != 1 {
		t.Errorf("Expected 1 transaction in 2022.12, but got: %d", len(dec))
	}

	nov, novExist := grouped[MonthDate{Year: 2022, Month: 11}]
	if !novExist {
		t.Error("Expected existance of 2022.11 in grouped transaction")
	}
	if len(nov) != 1 {
		t.Errorf("Expected 1 transaction in 2022.11, but got: %d", len(nov))
	}
}

func TestParseMonthDatePositive(t *testing.T) {
	testCases := map[string]MonthDate{
		"2023-01-20":      {Year: 2023, Month: 1},
		"2020-12-":        {Year: 2020, Month: 12},
		"2023-01":         {Year: 2023, Month: 1},
		"1999-09-22":      {Year: 1999, Month: 9},
		"2023-01-20-crap": {Year: 2023, Month: 1},
	}
	for input, expected := range testCases {
		testRes, tErr := parseMonthDate(input)
		if tErr != nil {
			t.Errorf("parseMonthDate error: %s", tErr.Error())
		}
		if testRes != expected {
			t.Errorf("expected parsing %s into [%s], but got [%s]",
				input, expected.String(), testRes.String())
		}
	}
}

func TestParseMonthDateNegative(t *testing.T) {
	testCases := [...]string{
		"2022",
		"2022-",
		"test",
		"",
		"2022/12/01",
		"2022 12",
		"dami-an",
	}
	for _, input := range testCases {
		tRes, tErr := parseMonthDate(input)
		if tErr == nil {
			t.Errorf("expecting failure for parseMonthDate(%s) but got success (%s)",
				input, tRes.String())
		}
	}
}
