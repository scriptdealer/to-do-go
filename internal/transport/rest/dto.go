package rest

import "errors"

var errInvalidPatchRequest = errors.New("update data has empty values")

type apiResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type itemPatchRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

func (update *itemPatchRequest) Validate() error {
	if update.Title == "" || update.Description == "" {
		return errInvalidPatchRequest
	}
	return nil
}
