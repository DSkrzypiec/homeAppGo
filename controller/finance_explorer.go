package controller

import (
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/finance"
	"homeApp/front"
	"html/template"
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
	Tooltip   template.HTML
}

type MonthlyAggResults struct {
	Phrase           string
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
		transactions, dErr = fw.DbClient.FinTransByDescription("%%")
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
	agg := finance.AggregateMonthly(transactions, "PLN")
	hist := finance.AggregatesMonthlyToChart(agg)

	chartData := make([]SingleMonthAgg, len(hist))
	for idx, agg := range hist {
		sma := SingleMonthAgg{
			MonthStr:  agg.MonthStr,
			AggValue:  agg.AggValueScaled,
			DataLabel: agg.DataLabel,
			Tooltip:   template.HTML(agg.Tooltip),
		}
		chartData[idx] = sma
	}

	tmplData := MonthlyAggResults{
		Phrase:           transactionsFilter,
		MonthlyChartData: chartData,
	}

	tmpl := front.FinanceExplorer()
	tmpl.Execute(w, tmplData)
}
