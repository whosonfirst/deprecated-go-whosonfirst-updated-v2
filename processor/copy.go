package processor

import (
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"log"
)

type CopyProcessor struct {
	Processor
	reader reader.Reader
	// writer writer.Writer
}

func NewCopyProcessor() (Processor, error) {

	r, err := reader.NewNullReader()

	if err != nil {
		return nil, err
	}
	
	// w := reader.NewNullWriter()

	d := CopyProcessor{
		reader: r,
		// writer: w,
	}

	return &d, nil
}

func (d *CopyProcessor) Name() error {
     return "copy"
}

func (d *CopyProcessor) Flush() error {
     return nil
}

// this received a CSV blob containing rows of commit_hash, repo, path

func (d *CopyProcessor) ProcessTask(task updated.Task) error {

	r, err := reader.NewGitHubReader(task.Repo)

	if err != nil {
		return err
	}

	for _, path := range task.Commits {
	
		fh, err := r.Read(path)

		if err != nil {
			return err
		}

		log.Println(fh)
	}

	return nil
}
