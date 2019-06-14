package funclown

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
)

type tableDTO interface {
	TableName() string
}

type queries interface {
	Find(tableDTO, ...OptFn) error
	FindMulti([]tableDTO, ...OptFn) error
}

type commands interface {
	Save(tableDTO, ...OptFn) error
	Delete(tableDTO, ...OptFn) error
}

// NewWriter provides interface for new crud interface.
type NewWriter interface {
	New(context.Context) WriterTxn
}

// Writer provides interface for repository use.
type Writer interface {
	queries
	commands
}

// WriterTxn provides interface for database operation with beginning transaction.
type WriterTxn interface {
	queries
	commands
	Begin()
}

// WriterInTxn provides interface for database operation in transaction.
type WriterInTxn interface {
	queries
	commands
	Commit()
	Rollback()
}

// ReadOnly provides read methods only.
type ReadOnly interface {
	queries
}

// NewReader provides interface for new finder interface.
type NewReader interface {
	New(context.Context) ReadOnly
}

// DBSelector provides read-only orm and writable orm.
type DBSelector interface {
	Reader(context.Context) ReadOnly
	Writer(context.Context) WriterTxn
}

// Factory provides building Wrapper object.
type Factory struct {
	master   *gorm.DB
	slave    *gorm.DB
	injector Injector
}

// Injector provides filling process in gorm.DB object or injecting process in other object by gorm.DB.
type Injector func(*gorm.DB) *gorm.DB

// NewFactory returns Wrapper with write permission.
func NewFactory(master, slave *gorm.DB, injectors ...Injector) *Factory {
	injector := func(db *gorm.DB) *gorm.DB {
		return db
	}
	if len(injectors) > 0 {
		injector = injectors[0]
	}
	return &Factory{master, slave, injector}
}

// Reader returns ORM Wrapper without write permission.
func (f *Factory) Reader(ctx context.Context) ReadOnly {
	return &Wrapper{db: f.injector(f.slave), state: txnAfter}
}

// Writer returns ORM Wrapper with write permission.
func (f *Factory) Writer(ctx context.Context) WriterTxn {
	return &Wrapper{db: f.injector(f.master), state: txnBefore}
}

type txnState int

const (
	txnBefore txnState = iota
	txnTerm
	txnAfter
)

// Wrapper provides being able to operate orm.DB with functional option pattern.
type Wrapper struct {
	db    *gorm.DB
	state txnState
}

// Begin provides delegation to gorm.DB.Begin().
func (w *Wrapper) Begin() {
	if w.state != txnBefore {
		panic(fmt.Errorf("current txnState:%d", w.state))
	}
	w.db = w.db.Begin()
	w.state = txnTerm
}

// Commit provides delegation to gorm.DB.Commit().
func (w *Wrapper) Commit() {
	if w.state != txnTerm {
		panic(fmt.Errorf("current txnState:%d", w.state))
	}
	w.db = w.db.Commit()
	w.state = txnAfter
}

// Rollback provides delegation to gorm.DB.Rollback().
func (w *Wrapper) Rollback() {
	switch w.state {
	case txnAfter:
		return
	case txnBefore:
		panic(fmt.Errorf("current txnState:%d", w.state))
	}
	w.db = w.db.Rollback()
	w.state = txnAfter
}

// Find provides wrapping operation for gorm.DB.First().
func (w *Wrapper) Find(out tableDTO, opts ...OptFn) error {
	sth := w.db
	for _, fn := range opts {
		sth = fn(sth)
	}
	res := sth.First(out)
	if res.RecordNotFound() {
		return ErrResourceNotFound
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// FindMulti provides wrapping operation for gorm.DB.Find().
func (w *Wrapper) FindMulti(outs []tableDTO, opts ...OptFn) error {
	sth := w.db
	for _, fn := range opts {
		sth = fn(sth)
	}
	res := sth.Find(outs)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// Save provides wrapping operation for gorm.DB.Save().
func (w *Wrapper) Save(resource tableDTO, opts ...OptFn) error {
	sth := w.db
	for _, fn := range opts {
		sth = fn(sth)
	}
	res := sth.Save(resource)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// Delete provides wrapping operation for gorm.DB.Delete().
func (w *Wrapper) Delete(value tableDTO, opts ...OptFn) error {
	sth := w.db
	for _, fn := range opts {
		sth = fn(sth)
	}
	res := sth.Delete(value)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
