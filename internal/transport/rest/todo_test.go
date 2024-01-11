package rest_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/scriptdealer/to-do-go/internal/services"
	"github.com/scriptdealer/to-do-go/internal/storage"
	"github.com/scriptdealer/to-do-go/internal/transport/rest"
	"github.com/stretchr/testify/suite"
)

type RouterSuite struct {
	suite.Suite

	db     *storage.InMemoryStorage
	logger *slog.Logger
	api    *rest.RESTful
}

func TestRouter(t *testing.T) {
	suite.Run(t, &RouterSuite{})
}

func (s *RouterSuite) SetupTest() {
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	s.db = storage.NewMemoryStorage(s.logger)
	services := services.NewComposite(
		s.db, s.logger,
		services.NewToDoService(s.db, s.logger),
	)
	s.api = rest.Init(services)

	w := httptest.NewRecorder()
	req := rest.TodoPatchRequest{Title: "1st", Description: "first test"}
	body, err := json.Marshal(req)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPost, "/todo", bytes.NewReader(body))

	s.api.AddItem(w, r)
	s.Equal(http.StatusOK, w.Code)
}

func (s *RouterSuite) TestAddItem_Ok() {
	w := httptest.NewRecorder()
	req := rest.TodoPatchRequest{Title: "2nd", Description: "second one"}
	body, err := json.Marshal(req)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPost, "/todo", bytes.NewReader(body))

	s.api.AddItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":true}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestAddItem_BadRequest() {
	w := httptest.NewRecorder()
	req := rest.TodoPatchRequest{Title: "test", Done: true}
	body, err := json.Marshal(req)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPost, "/todo", bytes.NewReader(body))

	s.api.AddItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":false,"error":"update data has empty values"}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestGetOne_Ok() {
	vars := map[string]string{"id": "1"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/todo/{id}", nil)
	r = mux.SetURLVars(r, vars)

	s.api.GetItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":true,"data":{"id":1,"title":"1st","description":"first test","done":false}}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestGetOne_NonExistent() {
	vars := map[string]string{"id": "2"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/todo/{id}", nil)
	r = mux.SetURLVars(r, vars)

	s.api.GetItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":false,"error":"no such item in storage"}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestDeletion_Unauthorized() {
	vars := map[string]string{"id": "1"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/todo/{id}", nil)
	r = mux.SetURLVars(r, vars)

	s.api.DeleteItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":false,"error":"not authorized"}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestDeletion_Ok() {
	vars := map[string]string{"id": "1"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/todo/{id}", nil)
	r = mux.SetURLVars(r, vars)
	r.Header.Add("Authorization", "Bearer kc74RbhOwtvVRcJhhJKpuDxSLwJY6oSC0iCfTJ2FsG0=")

	s.api.DeleteItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":true}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestDeletion_NonExistent() {
	vars := map[string]string{"id": "3"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/todo/{id}", nil)
	r = mux.SetURLVars(r, vars)
	r.Header.Add("Authorization", "Bearer kc74RbhOwtvVRcJhhJKpuDxSLwJY6oSC0iCfTJ2FsG0=")

	s.api.DeleteItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":false,"error":"no such item in storage"}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestUpdate_AllCases() {
	vars := map[string]string{"id": "3"}
	w := httptest.NewRecorder()
	req := rest.TodoPatchRequest{Title: "1st update!", Description: "updated"}
	body, err := json.Marshal(req)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPatch, "/todo/{id}", bytes.NewReader(body))
	r = mux.SetURLVars(r, vars)

	s.api.UpdateItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":false,"error":"no such item in storage"}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())

	vars["id"] = "1"
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodPatch, "/todo/{id}", bytes.NewReader(body))
	r2 = mux.SetURLVars(r2, vars)

	s.api.UpdateItem(w2, r2)
	s.Equal(http.StatusOK, w2.Code)
	expected = []byte(`{"success":true}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w2.Body.Bytes())

	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest(http.MethodGet, "/todo", nil)

	s.api.AllItems(w3, r3)
	s.Equal(http.StatusOK, w3.Code)
	expected = []byte(`{"success":true,"data":[{"id":1,"title":"1st update!","description":"updated","done":false}]}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w3.Body.Bytes())
}

func (s *RouterSuite) TestFilterByStatus_Ok() {
	vars := map[string]string{"selector": "active"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/todo/status/{selector}", nil)
	r = mux.SetURLVars(r, vars)

	s.api.FilterByStatus(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":true,"data":[{"id":1,"title":"1st","description":"first test","done":false}]}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())

	w2 := httptest.NewRecorder()
	req := rest.TodoPatchRequest{Title: "one more", Description: "Done one", Done: true}
	body, err := json.Marshal(req)
	s.Nil(err)
	r2 := httptest.NewRequest(http.MethodPost, "/todo", bytes.NewReader(body))
	s.api.AddItem(w2, r2)
	s.Equal(http.StatusOK, w2.Code)

	vars["selector"] = "done"
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest(http.MethodGet, "/todo/status/{selector}", nil)
	r3 = mux.SetURLVars(r3, vars)

	s.api.FilterByStatus(w3, r3)
	s.Equal(http.StatusOK, w.Code)
	expected = []byte(`{"success":true,"data":[{"id":2,"title":"one more","description":"Done one","done":true}]}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w3.Body.Bytes())
}
