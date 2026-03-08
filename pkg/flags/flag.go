package flags

import (
	"fmt"
)

type Flag struct {
	Cmd string
	Name string
	Args []string
	Description []string

	Handle func(args []string) error
}

func (f Flag) Usage() string {
	usage := f.Cmd + " "
	for _, arg := range f.Args {
		usage += fmt.Sprintf("<%s>", arg)
	}

	return usage
}
