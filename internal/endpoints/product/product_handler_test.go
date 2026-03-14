package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProductHandler(t *testing.T) {
	handler := NewProductHandler(nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.Repo)
}
