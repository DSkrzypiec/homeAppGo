package controller

import (
	"fmt"
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/finance"
	"homeApp/front"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	contrFinExPrefix = "controller/finEx"
)

type FinanceExplorer struct {
	DbClient       *db.Client
	TelegramClient *telegram.Client
	UserAuth       auth.UserAuthenticator
}

type SingleMonthAgg struct {
	MonthStr  string
	AggValue  float64
	DataLabel string
	Tooltip   string
}

type MonthlyAggResults struct {
	MonthlyChartData []SingleMonthAgg
}

// TODO
func (fw *FinanceExplorer) FinanceExplorerViewHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	transactionsFilter := r.FormValue("transactionsFilter")

	var transactions []db.BankTransaction
	var dErr error

	if transactionsFilter == "" {
		// No filtering
		// Load "all" (think about it) transactions from last 12 months and
		// aggregate
		transactions, dErr = fw.DbClient.FinTransByDescription("description") // TODO: change to no filtration later
		if dErr != nil {
			log.Error().Err(dErr).Msgf("[%s] cannot load transactions from database", contrFinExPrefix)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	} else {
		// Load filtered transactions from last 12 months and aggregate
		transactions, dErr = fw.DbClient.FinTransByDescription(transactionsFilter)
		if dErr != nil {
			log.Error().Err(dErr).Msgf("[%s] cannot load transactions from database for phrase [%s]",
				contrFinExPrefix, transactionsFilter)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		log.Info().Str("filter", contrFinExPrefix).Msgf("[%s] loading transactions in filtered version", contrFinExPrefix)
		// TODO...
	}
	fmt.Printf("Transactions: %v\n", transactions)
	//grouped := finance.GroupTransMonthly(transactions)
	agg := finance.AggregateMonthly(transactions, "PLN")

	for key, stats := range agg {
		fmt.Printf("[%s] - %v transactions\n", key.String(), stats)
	}

	tmpl := front.FinanceExplorer()
	tmpl.Execute(w, todoPreprMockData())
}

// TODO: This will be removed once result from Home DB will be integrated.
func todoPreprMockData() MonthlyAggResults {
	data := make([]SingleMonthAgg, 12)

	data[0] = sma("JUL 22", 0.456, "45zł", "TODO")
	data[1] = sma("AUG 22", 0.01, "0zł", "TODO")
	data[2] = sma("SEP 22", 0.76, "76zł", "TODO")
	data[3] = sma("OCT 22", 0.99, "99zł", "TODO")
	data[4] = sma("NOV 22", 0.15, "15zł", "TODO")
	data[5] = sma("DEC 23", 0.25, "25zł", "TODO")
	data[6] = sma("JAN 23", 0.50, "50zł", "TODO")
	data[7] = sma("FEB 23", 0.10, "10zł", "TODO")
	data[8] = sma("MAR 23", 0.87, "87zł", "TODO")
	data[9] = sma("APR 23", 0.75, "75zł", "TODO")
	data[10] = sma("MAY 23", 0.40, "40zł", "TODO")
	data[11] = sma("JUN 23", 0.60, "60zł", "TODO")

	return MonthlyAggResults{data}
}

func sma(month string, val float64, label string, tooltip string) SingleMonthAgg {
	return SingleMonthAgg{
		MonthStr:  month,
		AggValue:  val,
		DataLabel: label,
		Tooltip:   tooltip,
	}
}
