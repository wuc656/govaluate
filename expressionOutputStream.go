package govaluate

import (
	"strings"
)

/*
Holds a series of "transactions" which represent each token as it is output by an outputter (such as ToSQLQuery()).
Some outputs (such as SQL) require a function call or non-c-like syntax to represent an expression.
To accomplish this, this struct keeps track of each translated token as it is output, and can return and rollback those transactions.
*/
type expressionOutputStream struct {
	transactions []string
}

func (e *expressionOutputStream) add(transaction string) {
	e.transactions = append(e.transactions, transaction)
}

func (e *expressionOutputStream) rollback() string {

	index := len(e.transactions) - 1
	ret := e.transactions[index]

	e.transactions = e.transactions[:index]
	return ret
}

func (e *expressionOutputStream) createString(delimiter string) string {
	return strings.Join(e.transactions, delimiter)
}
