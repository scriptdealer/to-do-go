package services_test

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/scriptdealer/to-do-go/internal/services"
	"github.com/scriptdealer/to-do-go/internal/storage"
	"github.com/scriptdealer/to-do-go/known"
	"github.com/stretchr/testify/suite"
)

type TodoServiceSuite struct {
	suite.Suite

	logger   *slog.Logger
	db       *sql.DB
	mockedDB sqlmock.Sqlmock

	todo *services.TodoService
}

func TestTODO(t *testing.T) {
	suite.Run(t, &TodoServiceSuite{})
}

func (s *TodoServiceSuite) SetupTest() {
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, mock, err := sqlmock.New()
	s.NoError(err)
	s.db, s.mockedDB = db, mock

	s.todo = services.NewToDoService(
		&storage.PostgresStorage{Log: s.logger, DB: s.db},
		s.logger,
	)
}

func (s *TodoServiceSuite) TestCreate_Ok() {
	s.mockedDB.ExpectExec(regexp.QuoteMeta(`insert into todos (title, description, done) values ($1, $2, $3)`)).
		WithArgs("1st", "My first", true).WillReturnResult(sqlmock.NewResult(1, 1))
	err := s.todo.Create("1st", "My first", true)
	s.NoError(err)
}

func (s *TodoServiceSuite) TestCreate_DbFailure() {
	s.mockedDB.ExpectExec(regexp.QuoteMeta(`insert into todos (title, description, done) values ($1, $2, $3)`)).
		WithArgs("1st", "My first", true).WillReturnError(sql.ErrConnDone)

	err := s.todo.Create("1st", "My first", true)
	s.EqualError(err, "sql: connection is already closed")
}

func (s *TodoServiceSuite) TestGetOne_Ok() {
	mockedRows := sqlmock.NewRows([]string{"id", "title", "description", "done"}).AddRow(1, "1st", "My first", false)
	s.mockedDB.ExpectQuery(regexp.QuoteMeta(
		`select * from todos where id = $1`,
	)).WithArgs(1).WillReturnRows(mockedRows)

	got, err := s.todo.Get(1)
	s.NoError(err)
	s.Equal(&known.TodoItem{
		ID:          1,
		Title:       "1st",
		Description: "My first",
		Done:        false,
	}, got)
}

func (s *TodoServiceSuite) TestGetOne_NoRow() {
	emptyRows := sqlmock.NewRows([]string{"id", "title", "description", "done"})
	s.mockedDB.ExpectQuery(regexp.QuoteMeta(
		`select * from todos where id = $1`,
	)).WithArgs(2).WillReturnRows(emptyRows)

	got, err := s.todo.Get(2)
	s.EqualError(err, "no such item in storage")
	s.Nil(got)
}

func (s *TodoServiceSuite) TestGetOne_DbFailure() {
	s.mockedDB.ExpectQuery(regexp.QuoteMeta(
		`select * from todos where id = $1`,
	)).WithArgs(2).WillReturnError(sql.ErrConnDone)

	got, err := s.todo.Get(2)
	s.EqualError(err, "sql: connection is already closed")
	s.Nil(got)
}

func (s *TodoServiceSuite) TestGetAll_Ok() {
	mockedRows := sqlmock.NewRows([]string{"id", "title", "description", "done"}).
		AddRow(1, "1st", "My first", true).
		AddRow(2, "2nd", "My second", false)
	s.mockedDB.ExpectQuery(regexp.QuoteMeta(`select * from todos`)).WillReturnRows(mockedRows)

	got, err := s.todo.GetAll(context.Background())
	s.NoError(err)
	s.Equal(&known.TodoItem{
		ID:          1,
		Title:       "1st",
		Description: "My first",
		Done:        true,
	}, got[0])
	s.Equal(&known.TodoItem{
		ID:          2,
		Title:       "2nd",
		Description: "My second",
		Done:        false,
	}, got[1])
}

func (s *TodoServiceSuite) TestGetAll_CtxErr() {
	mockedRows := sqlmock.NewRows([]string{"id", "title", "description", "done"}).AddRow(1, "1st", "My first", true)
	s.mockedDB.ExpectQuery(regexp.QuoteMeta(`select * from todos`)).WillReturnRows(mockedRows)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.todo.GetAll(ctx)
	s.EqualError(err, context.Canceled.Error())

}

func (s *TodoServiceSuite) TestGetAll_ScanErr() {
	mockedRows := sqlmock.NewRows([]string{"id", "title", "description", "done"}).AddRow(1, nil, nil, true)
	s.mockedDB.ExpectQuery(regexp.QuoteMeta(`select * from todos`)).WillReturnRows(mockedRows)
	_, err := s.todo.GetAll(context.Background())
	s.EqualError(err, `sql: Scan error on column index 1, name "title": converting NULL to string is unsupported`)
}

func (s *TodoServiceSuite) TestUpdate_Ok() {
	s.mockedDB.ExpectExec(regexp.QuoteMeta(`update todos set title = $1, description = $2, done = $3 where id = $4`)).
		WithArgs("1st", "My first", true, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	err := s.todo.Update(1, "1st", "My first", true)
	s.NoError(err)
}

func (s *TodoServiceSuite) TestDelete_Ok() {
	s.mockedDB.ExpectExec(regexp.QuoteMeta(`delete from todos where id = $1`)).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	err := s.todo.Delete(1)
	s.NoError(err)
}
