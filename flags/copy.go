package flags

import (
	"strings"
)

type CopyFlags struct {
	Flags []string
}

func (fl *CopyFlags) String() string {
	return strings.Join(fl.Flags, "\n")
}

func (fl *CopyFlags) Set(value string) error {
	fl.Flags = append(fl.Flags, value)
	return nil
}
