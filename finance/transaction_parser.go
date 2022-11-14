package finance

// TransactionParser parses given file into slice of bank transactions.
type TransactionParser interface {
	Transactions(data []byte) ([]Transaction, error)
}

// Transaction represents single financial transaction.
type Transaction struct {
	AccountNumber         string
	ExecutionDate         string
	OrderDate             string
	Type                  *string
	AmountCurrency        string
	AmountValue           float64
	EndingBalanceCurrency *string
	EndingBalanceValue    *float64
	Description           string
}
