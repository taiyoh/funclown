package funclown

import "context"

type tableDTO interface {
	TableName() string
}

type tableDTOCollection interface{}

type queries interface {
	Count(tableDTO, ...OptFn) (int, error)
	Find(tableDTO, ...OptFn) error
	FindMulti(tableDTOCollection, ...OptFn) error
}

type commands interface {
	Save(tableDTO, ...OptFn) error
	Delete(tableDTO, ...OptFn) error
}

// TxnBehavior provides that this object can operate transaction.
type TxnBehavior interface {
	Begin()
	Rollback()
	Commit()
}

// Writer provides interface for repository use.
type Writer interface {
	queries
	commands
	TxnBehavior
}

// ReadOnly provides read methods only.
type ReadOnly interface {
	queries
}

// DBSelector provides read-only orm and writable orm.
type DBSelector interface {
	Reader(context.Context) ReadOnly
	Writer(context.Context) Writer
}
