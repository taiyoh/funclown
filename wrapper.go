package funclown

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
)

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

// Factory provides building Wrapper object.
type Factory struct {
	master   *gorm.DB
	slave    *gorm.DB
	injector InjectorFn
}

// Injector represents gorm.DB wrapper for injection.
type Injector struct {
	DB *gorm.DB
}

// InjectorFn provides filling process in gorm.DB object or injecting process in other object by gorm.DB.
type InjectorFn func(context.Context, *Injector)

// NewFactory returns Wrapper with write permission.
func NewFactory(master, slave *gorm.DB, injectors ...InjectorFn) *Factory {
	injector := func(context.Context, *Injector) {}
	if len(injectors) > 0 {
		injector = injectors[0]
	}
	return &Factory{master, slave, injector}
}

// Reader returns ORM Wrapper without write permission.
func (f *Factory) Reader(ctx context.Context) ReadOnly {
	return f.newWrapper(ctx, &Injector{f.slave}, txnAfter)
}

// Writer returns ORM Wrapper with write permission.
func (f *Factory) Writer(ctx context.Context) Writer {
	return f.newWrapper(ctx, &Injector{f.master}, txnBefore)
}

func (f *Factory) newWrapper(ctx context.Context, i *Injector, st txnState) *Wrapper {
	f.injector(ctx, i)
	return &Wrapper{db: i.DB, state: st}
}

// Wrapper provides being able to operate orm.DB with functional option pattern.
type Wrapper struct {
	db    *gorm.DB
	state txnState
}

// Begin provides delegation to gorm.DB.Begin().
func (w *Wrapper) Begin() {
	if w.state != txnBefore {
		panic(fmt.Errorf("current txnState: %s", w.state.String()))
	}
	w.db = w.db.Begin()
	w.state = txnTerm
}

// Commit provides delegation to gorm.DB.Commit().
func (w *Wrapper) Commit() {
	if w.state != txnTerm {
		panic(fmt.Errorf("current txnState: %s", w.state.String()))
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
		panic(fmt.Errorf("current txnState: %s", w.state.String()))
	}
	w.db = w.db.Rollback()
	w.state = txnAfter
}

// Count provides wrapping operation for gorm.DB.Count().
func (w *Wrapper) Count(dto tableDTO, opts ...OptFn) (int, error) {
	sth := w.db
	for _, fn := range opts {
		sth = fn(sth)
	}
	count := 0
	res := sth.Model(dto).Count(&count)
	if res.Error != nil {
		return count, res.Error
	}
	return count, nil
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
func (w *Wrapper) FindMulti(outs tableDTOCollection, opts ...OptFn) error {
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
