package processor

import (
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
)

type Processor interface {
	ProcessTask(task updated.Task) error
	Name() string
	Flush() error
}
