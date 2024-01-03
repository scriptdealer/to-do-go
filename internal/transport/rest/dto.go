package rest

import "errors"

var errInvalidPatchRequest = errors.New("update data has empty values")

type fetchResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ItemPatchRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

func (update *ItemPatchRequest) Validate() error {
	if update.Title == "" || update.Description == "" {
		return errInvalidPatchRequest
	}
	return nil
}
