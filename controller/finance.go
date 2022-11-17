package controller

import (
	"bytes"
	"errors"
	"fmt"
	"homeApp/auth"
	"homeApp/auth/telegram"
	"homeApp/db"
	"homeApp/finance"
	"homeApp/front"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	contrFinPrefix         = "controller/fin"
	maxTranactinosFileSize = int64(50 * 1024 * 1024) // 50 MiB
)

type Finance struct {
	DbClient       *db.Client
	TelegramClient *telegram.Client
	UserAuth       auth.UserAuthenticator
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

type FinanceUpload struct {
	Stats       *UploadStats
	UploadError *string
}

type UploadStats struct {
	NumOfTransactions int
	MinExecutionDate  string
	MaxExecutionDate  string
}

// FinanceViewHandler gets financial summary and execute finance main view
// template.
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

// FinanceInsertForm renders financial insert form for new files.
func (f *Finance) FinanceInsertForm(w http.ResponseWriter, r *http.Request) {
	tmpl := front.FinanceNewForm()
	execErr := tmpl.Execute(w, nil)
	if execErr != nil {
		log.Error().Err(execErr).Msgf("[%s] cannot render finance insert form", contrFinPrefix)
		http.Redirect(w, r, "/finance", http.StatusSeeOther)
	}
}

// FinanceUploadFile handles uploading new files with financial transactions
// into the database.
func (f *Finance) FinanceUploadFile(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start parsing new transactions form", contrFinPrefix)
	tmpl := front.FinanceNewForm()

	r.ParseMultipartForm(maxTranactinosFileSize)
	parserType := r.FormValue("parser-type")
	var buf bytes.Buffer
	file, _, err := r.FormFile("pkoFile")
	if err != nil {
		log.Error().Err(err).Msgf("[%s] couldn't get file from the form", contrDocPrefix)
		errDisplay := "Could not get file from the form, please retry"
		tmpl.Execute(w, FinanceUpload{UploadError: &errDisplay})
		return
	}
	defer file.Close()
	io.Copy(&buf, file)

	transParser, parserErr := parserTypeToParser(parserType)
	if parserErr != nil {
		log.Error().Err(parserErr).Msgf("[%s] parser selection failed", contrFinPrefix)
		errDisplay := "Could not parse given file. Please check if file is in correct format."
		tmpl.Execute(w, FinanceUpload{UploadError: &errDisplay})
		return
	}

	transactions, tErr := transParser.Transactions(buf.Bytes())
	if tErr != nil {
		log.Error().Err(tErr).Msgf("[%s] couldn't parse file into list of transactions", contrFinPrefix)
		errDisplay := "Could not parse given file. Please check if file is in correct format."
		tmpl.Execute(w, FinanceUpload{UploadError: &errDisplay})
		return
	}

	dbErr := f.DbClient.FinInsertTransactions(TransactionsToDbTransactions(transactions))
	if dbErr != nil {
		log.Error().Err(dbErr).Msgf("[%s] couldn't insert transactions into database", contrFinPrefix)
		errDisplay := "Insertion into database failed, please contact administrator"
		tmpl.Execute(w, FinanceUpload{UploadError: &errDisplay})
		return
	}

	uploadStats := prepUploadStats(transactions)
	stats := FinanceUpload{Stats: &uploadStats}

	msg := fmt.Sprintf("Uploaded %d financial transactions from %s to %s.",
		uploadStats.NumOfTransactions, uploadStats.MinExecutionDate, uploadStats.MaxExecutionDate)
	teleErr := SendTelegramMsgForUser(r, f.UserAuth, f.TelegramClient, f.DbClient, msg)
	if teleErr != nil {
		log.Error().Err(teleErr).Msgf("[%s] sending message to Telegram failed", contrFinPrefix)
	}

	log.Info().Dur("duration", time.Since(startTs)).Msgf("[%s] finished parsing new transactions form", contrFinPrefix)
	tmpl.Execute(w, stats)
}

func parserTypeToParser(ptype string) (finance.TransactionParser, error) {
	if ptype == "pkoxml" {
		return finance.PkoBankXmlParser{}, nil
	}
	return nil, errors.New("unsupported parser")
}

func prepUploadStats(newTransactions []finance.Transaction) UploadStats {
	minExecDate := "2900-01-01"
	maxExecDate := "1900-01-01"

	for _, t := range newTransactions {
		if t.ExecutionDate < minExecDate {
			minExecDate = t.ExecutionDate
		}
		if t.ExecutionDate > maxExecDate {
			maxExecDate = t.ExecutionDate
		}
	}

	return UploadStats{
		NumOfTransactions: len(newTransactions),
		MinExecutionDate:  minExecDate,
		MaxExecutionDate:  maxExecDate,
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

func TransactionsToDbTransactions(tt []finance.Transaction) []db.BankTransaction {
	dbT := make([]db.BankTransaction, len(tt))

	for i := 0; i < len(tt); i++ {
		dbT[i] = db.BankTransaction{
			AccountNumber:         tt[i].AccountNumber,
			ExecutionDate:         tt[i].ExecutionDate,
			OrderDate:             tt[i].OrderDate,
			Type:                  tt[i].Type,
			AmountCurrency:        tt[i].AmountCurrency,
			AmountValue:           tt[i].AmountValue,
			EndingBalanceCurrency: tt[i].EndingBalanceCurrency,
			EndingBalanceValue:    tt[i].EndingBalanceValue,
			Description:           tt[i].Description,
		}
	}
	return dbT
}
