package flags

import (
	"testing"

	"artifactor/pkg/flags"
	"github.com/stretchr/testify/assert"
)

func TestInitFlagRegistry(t *testing.T) {
	InitFlagRegistry()
	assert.NotNil(t, Flags)
}

func TestRegisterFlag(t *testing.T) {
	InitFlagRegistry()

	testFlag := flags.Flag{
		Cmd:         "--test-flag",
		Name:        "test-flag",
		Args:        []string{"arg1"},
		Description: []string{"Test flag description"},
		Handle: func(args []string) error {
			return nil
		},
	}

	RegisterFlag(testFlag)

	assert.Contains(t, Flags, "--test-flag")
}

func TestGetFlag(t *testing.T) {
	InitFlagRegistry()

	testFlag := flags.Flag{
		Cmd:         "--get-test",
		Name:        "get-test",
		Args:        []string{},
		Description: []string{},
		Handle: func(args []string) error {
			return nil
		},
	}

	RegisterFlag(testFlag)

	retrieved, err := GetFlag("--get-test")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "--get-test", retrieved.Cmd)
}

func TestGetFlag_NotFound(t *testing.T) {
	InitFlagRegistry()

	_, err := GetFlag("--nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInitToken(t *testing.T) {
	flag := InitToken(nil)
	assert.Equal(t, "--init-admin-token", flag.Cmd)
	assert.Equal(t, "init-admin-token", flag.Name)
	assert.NotNil(t, flag.Handle)
}
