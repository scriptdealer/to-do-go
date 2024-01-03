package storage

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/scriptdealer/to-do-go/known"
)

var errNoItem = errors.New("no such item in storage")

type ToDoStore interface {
	GetOne(id int) (*known.TodoItem, error)
	GetAll(ctx context.Context) ([]*known.TodoItem, error)
	Create(item *known.TodoItem) (*known.TodoItem, error)
	Update(item *known.TodoItem) (*known.TodoItem, error)
	Delete(id int) error
}

type InMemoryStorage struct {
	ram          map[int]known.TodoItem
	ramLock      sync.Mutex
	currentIndex int
	logger       *slog.Logger
}

func NewMemoryStorage(logger *slog.Logger) *InMemoryStorage {
	logger.Info("In-memory storage selected")

	return &InMemoryStorage{
		ram:    make(map[int]known.TodoItem),
		logger: logger,
	}
}

func (tds *InMemoryStorage) GetOne(id int) (*known.TodoItem, error) {
	tds.ramLock.Lock()
	defer tds.ramLock.Unlock()

	result, found := tds.ram[id]
	if found {
		return &result, nil
	}

	return nil, errNoItem
}

func (tds *InMemoryStorage) GetAll(ctx context.Context) ([]*known.TodoItem, error) {
	tds.ramLock.Lock()
	defer tds.ramLock.Unlock()

	result := make([]*known.TodoItem, 0)
	for k := range tds.ram {
		v := tds.ram[k]
		result = append(result, &v)
	}

	return result, nil
}

func (tds *InMemoryStorage) Create(item *known.TodoItem) (*known.TodoItem, error) {
	tds.ramLock.Lock()
	defer tds.ramLock.Unlock()

	tds.currentIndex++
	item.ID = tds.currentIndex
	tds.ram[tds.currentIndex] = *item

	return item, nil
}

func (tds *InMemoryStorage) Update(item *known.TodoItem) (*known.TodoItem, error) {
	tds.ramLock.Lock()
	defer tds.ramLock.Unlock()
	_, found := tds.ram[item.ID]
	if found {
		tds.ram[item.ID] = *item
		return item, nil
	}

	return nil, errNoItem
}

func (tds *InMemoryStorage) Delete(id int) error {
	tds.ramLock.Lock()
	defer tds.ramLock.Unlock()
	_, found := tds.ram[id]
	if found {
		delete(tds.ram, id)
		return nil
	}

	return errNoItem
}
