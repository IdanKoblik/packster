package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	origLog := Log
	defer func() { Log = origLog }()

	SetupLogger()
	assert.NotNil(t, Log)
}
