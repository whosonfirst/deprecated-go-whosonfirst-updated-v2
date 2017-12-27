package processor

import (
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"	
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
	"log"
)

type CopyProcessor struct {
	Processor
	reader reader.Reader
	writer writer.Writer	// please for to add multi writer support
}

func NewCopyProcessor() (Processor, error) {

	r, err := reader.NewNullReader()

	if err != nil {
		return nil, err
	}

	w, err := writer.NewNullWriter()

	if err != nil {
		return nil, err
	}

	pr := CopyProcessor{
		reader: r,
		writer: w,
	}

	return &pr, nil
}

func (pr *CopyProcessor) Name() string {
	return "copy"
}

func (pr *CopyProcessor) Flush() error {
	return nil
}

// this received a CSV blob containing rows of commit_hash, repo, path

func (pr *CopyProcessor) ProcessTask(task updated.Task) error {

	_, err := reader.NewGitHubReader(task.Repo)

	if err != nil {
		return err
	}

	for _, path := range task.Commits {

		log.Println("COPY", path)
		
		fh, err := pr.reader.Read(path)

		if err != nil {
			return err
		}
		
		err = pr.writer.Write(path, fh)
		
		if err != nil {
			return err
		}
	}

	return nil
}
