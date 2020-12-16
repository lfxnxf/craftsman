package sql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewMock(t *testing.T) {
	mock, closeFn, err := NewMock("test")
	if err != nil {
		return
	}
	defer closeFn()
	// mock sql
	rows := sqlmock.NewRows([]string{"id", "title", "body"}).
		AddRow(1, "post1", "hello").
		AddRow(2, "post2", "world")

	mock.ExpectQuery("^SELECT (.+) FROM `posts`").WillReturnRows(rows)
	// buz operation
	db := Get("test").Master()
	type Post struct {
		ID    int
		Title string
		Body  string
	}
	var p Post
	err = db.First(&p, 1).Error
	assert.Nil(t, err)
	assert.Equal(t, p, Post{ID: 2, Title: "post2", Body: "world"})
}
func TestNewMockGroup(t *testing.T) {
	group, mock, close, err := NewMockGroup()
	if err != nil {
		return
	}
	defer close()
	SQLGroupManager.Add("test", group)

	// mock sql
	rows := sqlmock.NewRows([]string{"id", "title", "body"}).
		AddRow(1, "post1", "hello").
		AddRow(2, "post2", "world")

	mock.ExpectQuery("^SELECT (.+) FROM `posts`").WillReturnRows(rows)
	// buz operation
	db := Get("test").Master()
	type Post struct {
		ID    int
		Title string
		Body  string
	}
	var p Post
	err = db.First(&p, 1).Error
	assert.Nil(t, err)
	assert.Equal(t, p, Post{ID: 2, Title: "post2", Body: "world"})

}
