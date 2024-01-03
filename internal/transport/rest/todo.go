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

func AllItems(w http.ResponseWriter, r *http.Request) {
	defer LogRecover()
	todos, err := serviceLayer.ToDos.GetAll(r.Context())
	serviceLayer.Log.Info("serving AllItems", slog.Int("count", len(todos)))
	respondWith(w, todos, err)
}

func GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	todo, err := serviceLayer.ToDos.Get(id)
	respondWith(w, todo, err)
}

func FilterByStatus(w http.ResponseWriter, r *http.Request) {
	todos, err := serviceLayer.ToDos.GetAll(r.Context())
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
	serviceLayer.Log.Info("serving filtered items", slog.Int("count", len(result)))
	respondWith(w, result, err)
}

func AddItem(w http.ResponseWriter, r *http.Request) {
	var (
		data itemPatchRequest
		todo *known.TodoItem
	)
	reqBody, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(reqBody, &data)
	serviceLayer.Log.Info("adding item", slog.String("body", fmt.Sprintf("%+v", data)))

	err := data.Validate()
	if err == nil {
		todo, err = serviceLayer.ToDos.Create(data.Title, data.Description)
	}
	respondWith(w, todo, err)
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	var (
		data itemPatchRequest
		todo *known.TodoItem
	)
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	err := json.NewDecoder(r.Body).Decode(&data)
	if err == nil {
		todo, err = serviceLayer.ToDos.Update(id, data.Title, data.Description, data.Done)
	}
	respondWith(w, todo, err)
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	if ok := authCheck(r); !ok {
		respondWith(w, nil, fmt.Errorf("not authorized"))
		return
	}
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	err := serviceLayer.ToDos.Delete(id)
	respondWith(w, nil, err)
}

func authCheck(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	authPrefix := "Bearer "

	if !strings.HasPrefix(authHeader, authPrefix) {
		return false
	}

	authKey := authHeader[len(authPrefix):]

	return authKey == staticAPIKey
}

func respondWith(w io.Writer, data any, err error) {
	reply := apiResponse{}
	if err == nil {
		reply.Success = true
		reply.Data = data
	} else {
		reply.Error = err.Error()
	}
	err = json.NewEncoder(w).Encode(reply)
	if err != nil {
		serviceLayer.Log.Warn("failed to respond", slog.String("reason", err.Error()))
	}
}
