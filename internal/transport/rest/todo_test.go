package rest

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
	"github.com/stretchr/testify/suite"
)

type RouterSuite struct {
	suite.Suite

	db     *storage.InMemoryStorage
	logger *slog.Logger
	router http.Handler
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
	s.router = InitHandlers(services)

	w := httptest.NewRecorder()
	req := itemPatchRequest{Title: "1st", Description: "first test"}
	body, err := json.Marshal(req)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPost, "/todo", bytes.NewReader(body))

	AddItem(w, r)
	s.Equal(http.StatusOK, w.Code)
}

func (s *RouterSuite) TestAddItem_Ok() {
	w := httptest.NewRecorder()
	req := itemPatchRequest{Title: "2nd", Description: "second one"}
	body, err := json.Marshal(req)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPost, "/todo", bytes.NewReader(body))

	AddItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":true,"data":{"id":2,"title":"2nd","description":"second one","done":false}}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestAddItem_BadRequest() {
	w := httptest.NewRecorder()
	req := itemPatchRequest{Title: "test", Done: true}
	body, err := json.Marshal(req)
	s.Nil(err)
	r := httptest.NewRequest(http.MethodPost, "/todo", bytes.NewReader(body))

	AddItem(w, r)
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

	GetItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":true,"data":{"id":1,"title":"1st","description":"first test","done":false}}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}

func (s *RouterSuite) TestDeletion_Unauthorized() {
	vars := map[string]string{"id": "1"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/todo/{id}", nil)
	r = mux.SetURLVars(r, vars)

	DeleteItem(w, r)
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

	DeleteItem(w, r)
	s.Equal(http.StatusOK, w.Code)
	expected := []byte(`{"success":true}`)
	expected = append(expected, 0xa)
	s.Equal(expected, w.Body.Bytes())
}
