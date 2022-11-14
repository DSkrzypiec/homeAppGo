package finance

import (
	"testing"
)

func TestSinglePkoTransaction(t *testing.T) {
	parser := PkoBankXmlParser{}
	transactions, err := parser.Transactions([]byte(singlePkoTransactionXml()))

	if err != nil {
		t.Errorf("couldn't parse PKO transactions: %v", err)
		return
	}

	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got: %d", len(transactions))
		return
	}

	t1 := transactions[0]
	const accNumber = "3310205356xxxx201986397"
	if t1.AccountNumber != accNumber {
		t.Errorf("expected %s account number, got: %s", accNumber, transactions[0].AccountNumber)
		return
	}

	execDate := "2022-04-30"
	if t1.ExecutionDate != execDate {
		t.Errorf("expected %s execution date, got: %s", execDate, t1.ExecutionDate)
		return
	}

	orderDate := "2022-04-29"
	if t1.OrderDate != orderDate {
		t.Errorf("expected %s order date, got: %s", orderDate, t1.OrderDate)
		return
	}

	typeStr := "Płatność kartą"
	if *t1.Type != typeStr {
		t.Errorf("expected %s type, got: %s", typeStr, *t1.Type)
		return
	}

	desc := `Tytuł: 000527242 75272420119124073189960
Numer karty: 535470******6202`
	if t1.Description != desc {
		t.Errorf("expected [%s] description, got: [%s]", desc, t1.Description)
		return
	}

	amount := -38.00
	if t1.AmountValue != amount {
		t.Errorf("expected %f amount, got %f", amount, t1.AmountValue)
		return
	}

	amountCurr := "PLN"
	if t1.AmountCurrency != amountCurr {
		t.Errorf("expected %s currency, got: %s", amountCurr, t1.AmountCurrency)
		return
	}

	balanceAmount := 14256.22
	if *t1.EndingBalanceValue != balanceAmount {
		t.Errorf("expected %f ending balance, got: %f", balanceAmount, *t1.EndingBalanceValue)
		return
	}

	balanceCurr := "PLN"
	if *t1.EndingBalanceCurrency != balanceCurr {
		t.Errorf("expected %s balance currency, got: %s", balanceCurr, *t1.EndingBalanceCurrency)
		return
	}
}

func TestTwoPkoTransactions(t *testing.T) {
	parser := PkoBankXmlParser{}
	transactions, err := parser.Transactions([]byte(twoPkoTransactionXml()))

	if err != nil {
		t.Errorf("couldn't parse PKO transactions: %v", err)
		return
	}

	if len(transactions) != 2 {
		t.Errorf("expected 2 transaction, got: %d", len(transactions))
		return
	}

	t1 := transactions[0]
	t2 := transactions[1]

	const accNumber = "3310205356xxxx201986397"
	if t1.AccountNumber != accNumber {
		t.Errorf("expected %s account number for t1, got: %s", accNumber, transactions[0].AccountNumber)
		return
	}
	if t2.AccountNumber != accNumber {
		t.Errorf("expected %s account number for t2, got: %s", accNumber, transactions[0].AccountNumber)
		return
	}

	if t1.Description != "desc1" {
		t.Errorf("expected desc1 description for t1, got: [%s]", t1.Description)
		return
	}
	if t2.Description != "desc2" {
		t.Errorf("expected desc2 description for t2, got: [%s]", t2.Description)
		return
	}

	amount1 := -38.00
	if t1.AmountValue != amount1 {
		t.Errorf("expected %f amount for t1, got %f", amount1, t1.AmountValue)
		return
	}
	amount2 := 138.00
	if t2.AmountValue != amount2 {
		t.Errorf("expected %f amount for t2, got %f", amount2, t2.AmountValue)
		return
	}

	balanceAmount1 := 14256.22
	if *t1.EndingBalanceValue != balanceAmount1 {
		t.Errorf("expected %f ending balance for t1, got: %f", balanceAmount1, *t1.EndingBalanceValue)
		return
	}
	balanceAmount2 := 15111.99
	if *t2.EndingBalanceValue != balanceAmount2 {
		t.Errorf("expected %f ending balance for t2, got: %f", balanceAmount2, *t2.EndingBalanceValue)
		return
	}
}

func singlePkoTransactionXml() string {
	return `<?xml version="1.0" encoding="utf-8"?>
<account-history>
<search>
  <account>3310205356xxxx201986397</account>
  <date since='2022-04-01' to='2022-04-30'/>
  <filtering>Wszystkie</filtering>
</search>
<operations>
  <operation>
    <exec-date>2022-04-30</exec-date>
    <order-date>2022-04-29</order-date>
    <type>Płatność kartą</type>
    <description>Tytuł: 000527242 75272420119124073189960
Numer karty: 535470******6202</description>
    <amount curr='PLN'>-38.00</amount><ending-balance curr='PLN'>+14256.22</ending-balance>
  </operation>
</operations>
</account-history>`
}

func twoPkoTransactionXml() string {
	return `<?xml version="1.0" encoding="utf-8"?>
<account-history>
<search>
  <account>3310205356xxxx201986397</account>
  <date since='2022-04-01' to='2022-05-01'/>
  <filtering>Wszystkie</filtering>
</search>
<operations>
  <operation>
    <exec-date>2022-04-30</exec-date>
    <order-date>2022-04-29</order-date>
    <type>type1</type>
    <description>desc1</description>
    <amount curr='PLN'>-38.00</amount><ending-balance curr='PLN'>+14256.22</ending-balance>
  </operation>
  <operation>
    <exec-date>2022-05-01</exec-date>
    <order-date>2022-04-30</order-date>
    <type>type2</type>
    <description>desc2</description>
    <amount curr='PLN'>+138.00</amount><ending-balance curr='PLN'>+15111.99</ending-balance>
  </operation>
</operations>
</account-history>`
}
