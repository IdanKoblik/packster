package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagUsage(t *testing.T) {
	tests := []struct {
		name     string
		flag     Flag
		expected string
	}{
		{
			name: "no args",
			flag: Flag{
				Cmd:  "--test",
				Args: []string{},
			},
			expected: "--test ",
		},
		{
			name: "single arg",
			flag: Flag{
				Cmd:  "--test",
				Args: []string{"arg1"},
			},
			expected: "--test <arg1>",
		},
		{
			name: "multiple args",
			flag: Flag{
				Cmd:  "--test",
				Args: []string{"arg1", "arg2", "arg3"},
			},
			expected: "--test <arg1><arg2><arg3>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.flag.Usage()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
