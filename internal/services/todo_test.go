package services

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type TodoServiceSuite struct {
	suite.Suite

	logger   *slog.Logger
	db       *sql.DB
	mockedDB sqlmock.Sqlmock
}

func TestTODO(t *testing.T) {
	suite.Run(t, &TodoServiceSuite{})
}

func (s *TodoServiceSuite) SetupTest() {
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, mock, err := sqlmock.New()
	s.NoError(err)
	s.db, s.mockedDB = db, mock
}
