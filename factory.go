package funclown

import (
	"context"

	"github.com/jinzhu/gorm"
)

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
