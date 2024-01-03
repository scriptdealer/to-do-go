package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	data := itemPatchRequest{Title: "test"}
	assert.EqualError(t, data.Validate(), errInvalidPatchRequest.Error())

	data.Description = data.Title
	data.Title = ""
	assert.EqualError(t, data.Validate(), errInvalidPatchRequest.Error())

	data.Title = data.Description
	assert.Nil(t, data.Validate())
}
