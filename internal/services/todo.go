package services

//go:generate mockgen -destination=todo_mock_test.go -source=todo.go -package=services_test

import (
	"context"
	"log/slog"

	"github.com/scriptdealer/to-do-go/internal/storage"
	"github.com/scriptdealer/to-do-go/known"
)

type TodoLogic interface {
	Create(title, description string, done bool) error
	Get(id int) (*known.TodoItem, error)
	GetAll(ctx context.Context) ([]*known.TodoItem, error)
	Update(id int, title, description string, done bool) error
	Delete(id int) error
}

type TodoService struct {
	store storage.ToDoStore
	Log   *slog.Logger
}

func NewToDoService(db storage.ToDoStore, logger *slog.Logger) *TodoService {
	return &TodoService{store: db, Log: logger}
}

func (tds *TodoService) Create(title, description string, done bool) error {
	item := known.TodoItem{
		Title:       title,
		Description: description,
		Done:        done,
	}
	return tds.store.Create(&item)
}

func (tds *TodoService) Update(id int, title, description string, done bool) error {
	patch := known.TodoItem{
		ID:          id,
		Title:       title,
		Description: description,
		Done:        done,
	}
	return tds.store.Update(&patch)
}

func (tds *TodoService) Delete(id int) error {
	return tds.store.Delete(id)
}

func (tds *TodoService) Get(id int) (*known.TodoItem, error) {
	return tds.store.GetOne(id)
}

func (tds *TodoService) GetAll(ctx context.Context) ([]*known.TodoItem, error) {
	return tds.store.GetAll(ctx)
}
