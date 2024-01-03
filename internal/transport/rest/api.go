package rest

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	"github.com/scriptdealer/to-do-go/internal/services"
)

var (
	serviceLayer *services.Composition
	staticAPIKey string
)

func InitHandlers(layer *services.Composition) http.Handler {
	serviceLayer = layer
	staticAPIKey = "kc74RbhOwtvVRcJhhJKpuDxSLwJY6oSC0iCfTJ2FsG0="
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/todo", AllItems).Methods(http.MethodGet)
	r.HandleFunc("/todo", AddItem).Methods(http.MethodPost)
	r.HandleFunc("/todo/{id}", GetItem).Methods(http.MethodGet)
	r.HandleFunc("/todo/{id}", UpdateItem).Methods(http.MethodPatch)
	r.HandleFunc("/todo/{id}", DeleteItem).Methods(http.MethodDelete)
	r.HandleFunc("/todo/status/{selector}", FilterByStatus).Methods(http.MethodGet)
	return r
}

func LogRecover() {
	if err := recover(); err != nil {
		serviceLayer.Log.Error("Fail:", slog.String("stack", string(debug.Stack())))
	}
}
