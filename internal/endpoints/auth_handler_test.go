package endpoints

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAuthHandler(t *testing.T) {
	handler := NewAuthHandler(nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.Repo)
}
