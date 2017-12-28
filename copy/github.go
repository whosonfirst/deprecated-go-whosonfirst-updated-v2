package copy

import (
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
	"log"
)

type GitHubCopier struct {
	Copier
	writer writer.Writer
}

func NewGitHubCopier(w writer.Writer) (Copier, error) {

	pr := GitHubCopier{
		writer: w,
	}

	return &pr, nil
}

func (pr *GitHubCopier) Name() string {
	return "whosonfirst-data"
}

func (pr *GitHubCopier) Flush() error {
	return nil
}

func (pr *GitHubCopier) ProcessTask(task updated.Task) error {

	r, err := reader.NewGitHubReader(task.Repo, "master")

	if err != nil {
		return err
	}

	for _, path := range task.Commits {

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

	return nil
}
