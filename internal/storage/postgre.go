package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/lib/pq" // driver import
	"github.com/scriptdealer/to-do-go/known"
)

type PostgreConfiguration struct {
	UserName string
	Password string
	Host     string
	Port     string
	DBName   string
}

func getConfig() *PostgreConfiguration {
	cfg := PostgreConfiguration{
		Host:     "db",
		Port:     "5432",
		DBName:   "todo_demo",
		UserName: "pguser",
		Password: "pgpassword",
	}

	host, found := os.LookupEnv("DB_HOST")
	if found {
		cfg.Host = host
	}

	user, found := os.LookupEnv("DB_USER")
	if found {
		cfg.UserName = user
	}

	pass, found := os.LookupEnv("DB_PASS")
	if found {
		cfg.Password = pass
	}

	return &cfg
}

type PostgresStorage struct {
	DB  *sql.DB
	cfg *PostgreConfiguration
	Log *slog.Logger
}

func NewPostgresStore(logger *slog.Logger) (*PostgresStorage, error) {
	logger.Info("Postgre storage selected")

	config := getConfig()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.UserName,
		config.Password,
		config.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Info("DB open failed", slog.String("reason", err.Error()))
		return nil, err
	}

	store := &PostgresStorage{
		DB:  db,
		cfg: config,
		Log: logger,
	}

	if err := store.Init(); err != nil {
		logger.Info("DB init failed", slog.String("reason", err.Error()))
		return nil, err
	}

	return store, nil
}

func (s *PostgresStorage) Init() error {
	return s.createToDoTable()
}

func (s *PostgresStorage) createToDoTable() error {
	query := `create table if not exists todos (
		id serial primary key,
		title varchar(100),
		description varchar(100),
		done boolean
	)`

	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresStorage) Create(item *known.TodoItem) error {
	query := `insert into todos (title, description, done) values ($1, $2, $3)`

	_, err := s.DB.Exec(
		query,
		item.Title,
		item.Description,
		item.Done,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) Update(item *known.TodoItem) error {
	_, err := s.DB.Exec(
		"update todos set title = $1, description = $2, done = $3 where id = $4",
		item.Title,
		item.Description,
		item.Done,
		item.ID,
	)
	return err
}

func (s *PostgresStorage) Delete(id int) error {
	_, err := s.DB.Exec("delete from todos where id = $1", id)
	return err
}

func (s *PostgresStorage) GetOne(id int) (*known.TodoItem, error) {
	rows, err := s.DB.Query("select * from todos where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanItem(rows)
	}

	return nil, errNoItem
}

func (s *PostgresStorage) GetAll(ctx context.Context) ([]*known.TodoItem, error) {
	rows, err := s.DB.QueryContext(ctx, "select * from todos")
	if err != nil {
		return nil, err
	}

	items := []*known.TodoItem{}
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func scanItem(rows *sql.Rows) (*known.TodoItem, error) {
	item := new(known.TodoItem)
	err := rows.Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Done)

	return item, err
}
