package rest

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	"github.com/scriptdealer/to-do-go/internal/services"
)

type RESTful struct {
	serviceLayer *services.Composition
	staticAPIKey string
	Router       http.Handler
}

func Init(layer *services.Composition) *RESTful {
	api := RESTful{
		serviceLayer: layer,
		staticAPIKey: "kc74RbhOwtvVRcJhhJKpuDxSLwJY6oSC0iCfTJ2FsG0=",
	}

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/todo", api.AllItems).Methods(http.MethodGet)
	r.HandleFunc("/todo", api.AddItem).Methods(http.MethodPost)
	r.HandleFunc("/todo/{id}", api.GetItem).Methods(http.MethodGet)
	r.HandleFunc("/todo/{id}", api.UpdateItem).Methods(http.MethodPatch)
	r.HandleFunc("/todo/{id}", api.DeleteItem).Methods(http.MethodDelete)
	r.HandleFunc("/todo/status/{selector}", api.FilterByStatus).Methods(http.MethodGet)

	api.Router = r

	return &api
}

func (r *RESTful) LogRecover() {
	if err := recover(); err != nil {
		r.serviceLayer.Log.Error("Fail:", slog.String("stack", string(debug.Stack())))
	}
}
