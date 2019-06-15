package funclown

type txnState int

//go:generate stringer -type=txnState
const (
	txnBefore txnState = iota
	txnTerm
	txnAfter
)
