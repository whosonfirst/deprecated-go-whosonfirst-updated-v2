package process

import (
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
)

type Processor interface {
	ProcessTask(updated.Task, chan string, chan error, chan bool)
	Name() string
	Flush() error
}
