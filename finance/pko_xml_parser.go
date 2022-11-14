package finance

import "encoding/xml"

type PkoBankXmlParser struct{}

// Transactions parses a slice of transactions from given XML file from PKO
// Bank (Polish bank).
func (p PkoBankXmlParser) Transactions(data []byte) ([]Transaction, error) {
	transactions := make([]Transaction, 0, 100)

	var rawTransactions pkoXmlTransactions
	xmlErr := xml.Unmarshal(data, &rawTransactions)
	if xmlErr != nil {
		return transactions, xmlErr
	}

	return rawTransactions.ToTransactions(), nil
}

func (pko pkoXmlTransactions) ToTransactions() []Transaction {
	transactions := make([]Transaction, 0, 100)
	account := pko.SearchInfo.Account

	for _, t := range pko.Transactions {
		tCopy := t
		transactions = append(transactions, Transaction{
			AccountNumber:         account,
			ExecutionDate:         t.ExecDate,
			OrderDate:             t.OrderDate,
			Type:                  &tCopy.TType,
			AmountCurrency:        t.Amount.Currency,
			AmountValue:           t.Amount.Amount,
			EndingBalanceCurrency: &tCopy.EndingBalance.Currency,
			EndingBalanceValue:    &tCopy.EndingBalance.Amount,
			Description:           t.Description,
		})
	}
	return transactions
}

type pkoXmlTransactions struct {
	XMLName      xml.Name            `xml:"account-history"`
	SearchInfo   pkoXmlSearch        `xml:"search"`
	Transactions []pkoXmlTransaction `xml:"operations>operation"`
}

type pkoXmlSearch struct {
	XMLName xml.Name `xml:"search"`
	Account string   `xml:"account"`
}

type pkoXmlTransaction struct {
	XMLName       xml.Name            `xml:"operation"`
	ExecDate      string              `xml:"exec-date"`
	OrderDate     string              `xml:"order-date"`
	TType         string              `xml:"type"`
	Description   string              `xml:"description"`
	Amount        pkoXmlAmount        `xml:"amount"`
	EndingBalance pkoXmlEndingBalance `xml:"ending-balance"`
}

type pkoXmlAmount struct {
	XMLName  xml.Name `xml:"amount"`
	Amount   float64  `xml:",chardata"`
	Currency string   `xml:"curr,attr"`
}

type pkoXmlEndingBalance struct {
	XMLName  xml.Name `xml:"ending-balance"`
	Amount   float64  `xml:",chardata"`
	Currency string   `xml:"curr,attr"`
}
