package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/scriptdealer/to-do-go/known"
)

func (rest *RESTful) AllItems(w http.ResponseWriter, r *http.Request) {
	defer rest.LogRecover()
	todos, err := rest.serviceLayer.ToDos.GetAll(r.Context())
	rest.serviceLayer.Log.Info("serving AllItems", slog.Int("count", len(todos)))
	rest.respondWith(w, todos, err)
}

func (rest *RESTful) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	todo, err := rest.serviceLayer.ToDos.Get(id)
	rest.respondWith(w, todo, err)
}

func (rest *RESTful) FilterByStatus(w http.ResponseWriter, r *http.Request) {
	todos, err := rest.serviceLayer.ToDos.GetAll(r.Context())
	vars := mux.Vars(r)
	status := vars["selector"]

	result := []known.TodoItem{}
	for i := range todos {
		if todos[i].Done && status == "done" {
			result = append(result, *todos[i])
		}

		if !todos[i].Done && status == "active" {
			result = append(result, *todos[i])
		}
	}
	rest.serviceLayer.Log.Info("serving filtered items", slog.Int("count", len(result)))
	rest.respondWith(w, result, err)
}

func (rest *RESTful) AddItem(w http.ResponseWriter, r *http.Request) {
	var data TodoPatchRequest
	reqBody, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(reqBody, &data)
	rest.serviceLayer.Log.Info("adding item", slog.String("body", fmt.Sprintf("%+v", data)))

	err := data.Validate()
	if err == nil {
		err = rest.serviceLayer.ToDos.Create(data.Title, data.Description, data.Done)
	}
	rest.respondWith(w, nil, err)
}

func (rest *RESTful) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var data TodoPatchRequest
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	err := json.NewDecoder(r.Body).Decode(&data)
	if err == nil {
		rest.serviceLayer.Log.Info("updating item", slog.Int("id", id), slog.String("with", fmt.Sprintf("%+v", data)))
		err = rest.serviceLayer.ToDos.Update(id, data.Title, data.Description, data.Done)
	}
	rest.respondWith(w, nil, err)
}

func (rest *RESTful) DeleteItem(w http.ResponseWriter, r *http.Request) {
	if ok := rest.authCheck(r); !ok {
		rest.respondWith(w, nil, fmt.Errorf("not authorized"))
		return
	}
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	rest.serviceLayer.Log.Info("deleting item", slog.Int("id", id))
	err := rest.serviceLayer.ToDos.Delete(id)
	rest.respondWith(w, nil, err)
}

func (rest *RESTful) authCheck(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	authPrefix := "Bearer "

	if !strings.HasPrefix(authHeader, authPrefix) {
		return false
	}

	authKey := authHeader[len(authPrefix):]

	return authKey == rest.staticAPIKey
}

func (rest *RESTful) respondWith(w io.Writer, data any, err error) {
	reply := apiResponse{}
	if err == nil {
		reply.Success = true
		reply.Data = data
	} else {
		reply.Error = err.Error()
	}
	err = json.NewEncoder(w).Encode(reply)
	if err != nil {
		rest.serviceLayer.Log.Warn("failed to respond", slog.String("reason", err.Error()))
	}
}
