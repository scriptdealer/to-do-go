package services

import (
	"database/sql"
	"log/slog"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/scriptdealer/to-do-go/internal/storage"
	"github.com/scriptdealer/to-do-go/known"
	"github.com/stretchr/testify/suite"
)

type TodoServiceSuite struct {
	suite.Suite

	logger   *slog.Logger
	db       *sql.DB
	mockedDB sqlmock.Sqlmock

	todo *TodoService
}

func TestTODO(t *testing.T) {
	suite.Run(t, &TodoServiceSuite{})
}

func (s *TodoServiceSuite) SetupTest() {
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, mock, err := sqlmock.New()
	s.NoError(err)
	s.db, s.mockedDB = db, mock

	s.todo = NewToDoService(
		&storage.PostgresStorage{Log: s.logger, DB: db},
		s.logger,
	)
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
