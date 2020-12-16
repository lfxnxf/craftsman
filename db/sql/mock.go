package sql

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

// NewMock returns sqlmock.Sqlmock and db.close() func.
// sqlmock.Sqlmock doc shows https://godoc.org/github.com/DATA-DOG/go-sqlmock
func NewMock(name string) (mock sqlmock.Sqlmock, closeFunc func(), err error) {
	group, mock, closeFunc, err := NewMockGroup()
	SQLGroupManager.Add(name, group)
	return
}

// NewMockGroup returns builtin mock group and sqlmock.
// sqlmock.Sqlmock can mock data for all SQL command
func NewMockGroup() (group *Group, mock sqlmock.Sqlmock, closeFunc func(), err error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return
	}
	closeFunc = func() {
		db.Close()
	}
	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		return
	}
	client := &Client{DB: gormDB}
	group = &Group{
		master:  client,
		replica: []*Client{client},
		next:    0,
		total:   1,
	}
	return
}
