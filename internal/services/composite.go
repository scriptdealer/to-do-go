package services

import (
	"log/slog"
	"os"

	"github.com/scriptdealer/to-do-go/internal/storage"
)

type Composition struct {
	// Config     *Configuration
	DB           storage.ToDoStore
	Interruption chan os.Signal
	Log          *slog.Logger
	ToDos        TodoLogic
	// Users        *userService
}

func NewComposite(db storage.ToDoStore, logger *slog.Logger, todos TodoLogic) *Composition {
	return &Composition{
		DB:           db,
		Log:          logger,
		ToDos:        todos,
		Interruption: make(chan os.Signal, 1),
	}
}
