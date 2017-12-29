package copy

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
	"sync"
)

type GitHubCopier struct {
	Copier
	writer  writer.Writer
	clients int
}

func NewGitHubCopier(w writer.Writer) (Copier, error) {

	pr := GitHubCopier{
		writer:  w,
		clients: 10,
	}

	return &pr, nil
}

func (pr *GitHubCopier) Name() string {
	return "whosonfirst-data"
}

func (pr *GitHubCopier) Flush() error {
	return nil
}

func (pr *GitHubCopier) ProcessTask(task updated.Task, status_ch chan string, error_ch chan error, done_ch chan bool) {

	defer func() {
		done_ch <- true
	}()

	r, err := reader.NewGitHubReader(task.Repo, "master")

	if err != nil {
		error_ch <- err
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	throttle_ch := make(chan bool, pr.clients)

	for i := 0; i < pr.clients; i++ {
		throttle_ch <- true
	}

	wg := new(sync.WaitGroup)

	count := len(task.Commits)

	for i, path := range task.Commits {

		<-throttle_ch

		wg.Add(1)

		go func(ctx context.Context, path string, i int) {

			defer func() {
				throttle_ch <- true
				wg.Done()
			}()

			select {

			case <-ctx.Done():
				return
			default:

				status_ch <- fmt.Sprintf("[%s] %d/%d read %s", pr.Name(), i, count, path)

				fh, err := r.Read(path)

				if err != nil {
					msg := fmt.Sprintf("[%s] %d/%d %s", pr.Name(), i, count, err.Error())
					error_ch <- errors.New(msg)

					cancel()
					return
				}

				status_ch <- fmt.Sprintf("[%s] %d/%d write %s", pr.Name(), i, count, path)

				err = pr.writer.Write(path, fh)

				fh.Close()

				if err != nil {
					msg := fmt.Sprintf("[%s] %d/%d %s", pr.Name(), i, count, err.Error())
					error_ch <- errors.New(msg)

					cancel()
					return
				}
			}

		}(ctx, path, i)
	}

	wg.Wait()
}
