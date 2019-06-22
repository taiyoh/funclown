package funclown

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

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
