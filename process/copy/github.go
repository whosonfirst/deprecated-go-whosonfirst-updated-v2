package copy

// this should not be considered a general purpose github copier yet - it
// is still largely whosonfirst-data specific (20180121/thisisaaronland)

import (
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
)

type GitHubCopier struct {
	Copier
	writer writer.Writer
}

func NewGitHubCopier(w writer.Writer) (Copier, error) {

	pr := GitHubCopier{
		writer: w,
		// we define reader on a per ProcessTask instance below
	}

	return &pr, nil
}

func (pr *GitHubCopier) Name() string {
	return "github-copy"
}

func (pr *GitHubCopier) Flush() error {
	return nil
}

func (pr *GitHubCopier) ProcessTask(t updated.Task, status_ch chan string, err_ch chan error, done_ch chan bool) {

	defer func() {
		done_ch <- true
	}()

	r, err := reader.NewGitHubReader(t.Repo, "master")

	if err != nil {
		err_ch <- err
		return
	}

	// this is defined in readwrite.go

	ReadWriteCopy(pr, r, pr.writer, t, status_ch, err_ch)
}
