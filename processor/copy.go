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

	/*
	cfg := writer.S3Config{
		Bucket: "data.whosonfirst.org",
		Prefix: "",
		Region: "us-east-1",
		Credentials: "whosonfirst",
	}
	
	w, err := writer.NewS3Writer(cfg)
	*/
	
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

	r, err := reader.NewGitHubReader(task.Repo, "master")

	if err != nil {
		return err
	}

	for _, path := range task.Commits {

		log.Println("COPY", path)
		
		fh, err := r.Read(path)

		if err != nil {
		   	log.Println("ERR", path, err)
			continue
		}
		
		err = pr.writer.Write(path, fh)

	      	fh.Close()
		
		if err != nil {
		   	log.Println("ERR", path, err)
			continue
		}
	}

	// log.Println("DONE TASK")
	return nil
}
