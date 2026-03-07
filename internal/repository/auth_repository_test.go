package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCacheKey(t *testing.T) {
	repo := &AuthRepository{}

	tests := []struct {
		token    string
		expected string
	}{
		{"test-token", "token:test-token"},
		{"abc123", "token:abc123"},
		{"", "token:"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			key := repo.getCacheKey(tt.token)
			assert.Equal(t, tt.expected, key)
		})
	}
}
