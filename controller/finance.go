package controller

import (
	"fmt"
	"homeApp/db"
	"homeApp/front"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

const contrFinPrefix = "controller/fin"

type Finance struct {
	DbClient *db.Client
}

type FinanceData struct {
	MonthlyAggregation []FinancialMonthlyAgg
}

type FinancialMonthlyAgg struct {
	YearMonth         string
	NumOfTransactions int
	Inflow            string
	Outflow           string
}

func (f *Finance) FinanceViewHandler(w http.ResponseWriter, r *http.Request) {
	monthlyAgg, err := f.getMonthlyAgg()
	if err != nil {
		log.Error().Err(err).Msgf("[%s] cannot perform monthly financial aggregartion", contrFinPrefix)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	tmpl := front.Finance()
	execErr := tmpl.Execute(w, FinanceData{monthlyAgg})
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render finance view", contrFinPrefix)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
}

func (f *Finance) FinanceInsertForm(w http.ResponseWriter, r *http.Request) {
	tmpl := front.FinanceNewForm()
	execErr := tmpl.Execute(w, nil)
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render finance insert form", contrFinPrefix)
		http.Redirect(w, r, "/finance", http.StatusSeeOther)
	}
}

func (f *Finance) FinanceUploadFile(w http.ResponseWriter, r *http.Request) {
	// TODO
	// Upload file, parse into transaction, upload into DB, send statistics
	// into /finance-new
	var stats string // TODO

	tmpl := front.FinanceNewForm()
	execErr := tmpl.Execute(w, stats)
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render finance insert form with stats", contrFinPrefix)
		http.Redirect(w, r, "/finance", http.StatusSeeOther)
	}
}

func (f *Finance) getMonthlyAgg() ([]FinancialMonthlyAgg, error) {
	const year = 2022 // TODO
	aggs := make([]FinancialMonthlyAgg, 0, 12)

	for month := 1; month <= 12; month++ {
		ts, dErr := f.DbClient.FinTransMonthly(2022, month)
		if dErr != nil {
			return aggs, dErr
		}
		monthStr := strconv.Itoa(month)
		if month < 10 {
			monthStr = "0" + monthStr
		}
		yearMonth := fmt.Sprintf("%d-%s", year, monthStr)
		aggs = append(aggs, transAgg(yearMonth, ts))
	}

	return aggs, nil
}

func transAgg(date string, trans []db.BankTransaction) FinancialMonthlyAgg {
	cnt := 0
	inflow := 0.0
	outflow := 0.0

	for _, t := range trans {
		cnt++
		if t.AmountValue > 0.0 {
			inflow += t.AmountValue
			continue
		}
		outflow += t.AmountValue
	}

	return FinancialMonthlyAgg{
		YearMonth:         date,
		NumOfTransactions: cnt,
		Inflow:            fmt.Sprintf("%.2f", inflow),
		Outflow:           fmt.Sprintf("%.2f", outflow),
	}
}
