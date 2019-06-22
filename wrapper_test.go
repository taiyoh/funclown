package funclown_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/taiyoh/funclown"
)

func initDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	gdb, err := gorm.Open("mysql", db)
	if err != nil {
		panic(err)
	}
	// gdb.LogMode(true)

	return gdb, mock, nil
}

const createdFormat = "2006-01-02 15:04:05"

type anyTime struct{}

func (anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type user struct {
	ID        uint       `gorm:"primary_key"`
	Name      string     `gorm:"column:name"`
	CreatedAt time.Time  `gorm:"datetime;column:created_at"`
	UpdatedAt time.Time  `gorm:"datetime;column:updated_at"`
	DeletedAt *time.Time `gorm:"datetime;column:deleted_at"`
}

func (user) TableName() string {
	return "users"
}

func newRefRows() *sqlmock.Rows {
	cols := []string{"id", "name", "created_at", "updated_at", "deleted_at"}
	return sqlmock.NewRows(cols)
}

func TestWrapperSave(t *testing.T) {
	gdb, mock, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer gdb.Close()
	factory := funclown.NewFactory(gdb, gdb)
	db := factory.Writer(context.Background())

	updateStmt := "UPDATE `users` SET `name` = ?, `created_at` = ?, `updated_at` = ?, `deleted_at` = ?  WHERE `users`.`id` = ?"
	findStmt := "SELECT * FROM `users` WHERE `users`.`deleted_at` IS NULL AND `users`.`id` = ? ORDER BY `users`.`id` ASC LIMIT 1"
	insertStmt := "INSERT INTO `users` (`id`,`name`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?)"

	mock.ExpectExec(regexp.QuoteMeta(updateStmt)).
		WithArgs("hoge", &anyTime{}, &anyTime{}, nil, 123).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(findStmt)).
		WithArgs(123).
		WillReturnRows(newRefRows())
	mock.ExpectExec(regexp.QuoteMeta(insertStmt)).
		WithArgs(123, "hoge", &anyTime{}, &anyTime{}, nil).
		WillReturnResult(sqlmock.NewResult(123, 1))

	now := time.Now().UTC()
	u := &user{ID: 123, Name: "hoge", CreatedAt: now, UpdatedAt: now}
	err = db.Save(u, funclown.IgnoreSoftDelete())
	if err != nil {
		t.Error(err)
	}
}

func TestWrapperCount(t *testing.T) {
	gdb, mock, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer gdb.Close()
	factory := funclown.NewFactory(gdb, gdb)
	db := factory.Reader(context.Background())

	countStmt := "SELECT count(*) FROM `users` WHERE `users`.`deleted_at` IS NULL AND ((id = ?))"

	mock.ExpectQuery(regexp.QuoteMeta(countStmt)).
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1000))

	count, err := db.Count(&user{}, funclown.Where("id = ?", 123))
	if err != nil {
		t.Fatal(err)
	}
	if count != 1000 {
		t.Error("wrong count", count)
	}
}

func TestWrapperFind(t *testing.T) {
	gdb, mock, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer gdb.Close()
	factory := funclown.NewFactory(gdb, gdb)
	db := factory.Reader(context.Background())

	now := time.Now().UTC()
	u := &user{}
	findStmt := "SELECT * FROM `users` WHERE `users`.`deleted_at` IS NULL AND ((id = ?)) ORDER BY `users`.`id` ASC LIMIT 1"

	mock.ExpectQuery(regexp.QuoteMeta(findStmt)).
		WithArgs(123).
		WillReturnRows(newRefRows())

	if err := db.Find(u, funclown.Where("id = ?", 123)); err != funclown.ErrResourceNotFound {
		t.Error(err)
	}
	if *u != (user{}) {
		t.Error("value is filled")
	}

	rows := newRefRows().AddRow(123, "hoge", now, now, nil)
	mock.ExpectQuery(regexp.QuoteMeta(findStmt)).
		WithArgs(123).
		WillReturnRows(rows)

	if err := db.Find(u, funclown.Where("id = ?", 123)); err != nil {
		t.Error(err)
	}
}

func TestWrapperFindWithLock(t *testing.T) {
	gdb, mock, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer gdb.Close()
	factory := funclown.NewFactory(gdb, gdb)
	db := factory.Reader(context.Background())

	now := time.Now().UTC()
	u := &user{}
	findStmt := "SELECT * FROM `users` WHERE (id = ?) ORDER BY `users`.`id` ASC LIMIT 1 FOR UPDATE"
	rows := newRefRows().AddRow(123, "hoge", now, now, nil)
	mock.ExpectQuery(regexp.QuoteMeta(findStmt)).
		WithArgs(123).
		WillReturnRows(rows)

	fns := funclown.OptFnCollection{}.
		Add(funclown.Where("id = ?", 123)).
		Add(funclown.IgnoreSoftDelete()).
		Add(funclown.ForUpdate())

	if err := db.Find(u, fns.Slice()...); err != nil {
		t.Error(err)
	}
}

func TestWrapperFindMulti(t *testing.T) {
	gdb, mock, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer gdb.Close()
	factory := funclown.NewFactory(gdb, gdb)
	db := factory.Reader(context.Background())

	now := time.Now().UTC()
	u := &user{}
	findStmt := "SELECT * FROM `users` WHERE (id in (?, ?)) ORDER BY id DESC LIMIT 10 OFFSET 3"
	rows := newRefRows().
		AddRow(123, "hoge", now, now, nil).
		AddRow(456, "fuga", now, now, nil)
	mock.ExpectQuery(regexp.QuoteMeta(findStmt)).
		WithArgs(123, 456).
		WillReturnRows(rows)

	fns := funclown.OptFnCollection{}.
		Add(funclown.Where("id in (?, ?)", 123, 456)).
		Add(funclown.Order("id DESC")).
		Add(funclown.Limit(10)).
		Add(funclown.Offset(3)).
		Add(funclown.IgnoreSoftDelete())

	if err := db.FindMulti(u, fns.Slice()...); err != nil {
		t.Error(err)
	}
}

func TestWrapperDelete(t *testing.T) {
	gdb, mock, err := initDB()
	if err != nil {
		t.Fatal(err)
	}
	defer gdb.Close()
	factory := funclown.NewFactory(gdb, gdb)
	db := factory.Writer(context.Background())

	t.Run("soft delete", func(t *testing.T) {
		u := &user{}
		deleteStmt := "UPDATE `users` SET `deleted_at`=? WHERE `users`.`deleted_at` IS NULL AND ((id = ?))"

		mock.ExpectExec(regexp.QuoteMeta(deleteStmt)).
			WithArgs(&anyTime{}, 123).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := db.Delete(u, funclown.Where("id = ?", 123)); err != nil {
			t.Error(err)
		}
	})

	t.Run("hard delete", func(t *testing.T) {
		u := &user{}
		deleteStmt := "DELETE FROM `users` WHERE (id = ?)"

		mock.ExpectExec(regexp.QuoteMeta(deleteStmt)).
			WithArgs(123).
			WillReturnResult(sqlmock.NewResult(0, 1))

		fns := funclown.OptFnCollection{}.
			Add(funclown.Where("id = ?", 123)).
			Add(funclown.IgnoreSoftDelete())

		if err := db.Delete(u, fns.Slice()...); err != nil {
			t.Error(err)
		}
	})

}
