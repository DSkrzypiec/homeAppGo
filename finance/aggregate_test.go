package finance

import "testing"

func TestParseMonthDatePositive(t *testing.T) {
	testCases := map[string]MonthDate{
		"2023-01-20":      {Year: 2023, Month: 1},
		"2020-12-":        {Year: 2020, Month: 12},
		"2023-01":         {Year: 2023, Month: 1},
		"1999-09-22":      {Year: 1999, Month: 9},
		"2023-01-20-crap": {Year: 2023, Month: 1},
	}
	for input, expected := range testCases {
		testRes, tErr := ParseMonthDate(input)
		if tErr != nil {
			t.Errorf("ParseMonthDate error: %s", tErr.Error())
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
		tRes, tErr := ParseMonthDate(input)
		if tErr == nil {
			t.Errorf("expecting failure for ParseMonthDate(%s) but got success (%s)",
				input, tRes.String())
		}
	}
}
