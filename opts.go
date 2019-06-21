package funclown

import "github.com/jinzhu/gorm"

// OptFn provides pattern function for database operation.
type OptFn func(db *gorm.DB) *gorm.DB

// Where returns function for orm.DB.Where operation.
func Where(placeholder string, arg interface{}, optargs ...interface{}) OptFn {
	whereArgs := make([]interface{}, 0, len(optargs)+1)
	whereArgs = append(whereArgs, arg)
	whereArgs = append(whereArgs, optargs...)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(placeholder, whereArgs...)
	}
}

// Order returns function for orm.DB.Order operation.
func Order(value interface{}, reorder ...bool) OptFn {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(value, reorder...)
	}
}

// Limit returns function for orm.DB.Limit operation.
func Limit(limit interface{}) OptFn {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// Offset returns function for orm.DB.Offset operation.
func Offset(offset interface{}) OptFn {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}

// ForUpdate returns function for orm.DB.Set("gorm:query_option", "for update") operation.
func ForUpdate() OptFn {
	return func(db *gorm.DB) *gorm.DB {
		return db.Set("gorm:query_option", "for update")
	}
}

// IgnoreSoftDelete provides unscoped flag. this makes selectable even if soft deleted.
func IgnoreSoftDelete() OptFn {
	return func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}
}

// OptFnCollection provides operation object for OptFn list.
type OptFnCollection []OptFn

// Add returns new OptFn list with given functions.
func (c OptFnCollection) Add(fns ...OptFn) OptFnCollection {
	if len(fns) <= 0 {
		return c
	}
	return append(c, fns...)
}
