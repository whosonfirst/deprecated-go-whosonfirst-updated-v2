package process

import (
	"context"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
)

type RWProcessor struct {
	writer writer.Writer	// remember this might be a MultiWriter
}

func NewRWProcessor() (Processor, error) {
	return nil, errors.New("Please write me")
}

func (pr *RWProcessor) ProcessTask(t updated.Task, what_ch chan string, err_ch chan error, done_ch chan bool) {

	defer func() {
		done_ch <- true
	}()

	err_ch <- errors.New("TOO SOON")
	return

	r, err := reader.NewGitHubReader(t.Repo, "master")

	if err != nil {
		err_ch <- err
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	local_done_ch := make(chan bool)
	local_err_ch := make(chan error)

	for _, path := range t.Commits {

		// this should almost certainly have a throttle...
		// (20180120/thisisaaronland)
		
		go func(ctx context.Context, path string, local_done_ch chan bool, local_err_ch chan error) {

			defer func() {
				local_done_ch <- true
			}()

			select {

			case <-ctx.Done():
				return
			default:
			
				fh, err := r.Read(path)

				if err != nil {
					local_err_ch <- err
					return
				}

				err = pr.writer.Write(path, fh)

				if err != nil {
					local_err_ch <- err
					return
				}

			}

		}(ctx, path, local_done_ch, local_err_ch)

	}

	remaining := len(t.Commits)

	for remaining > 0 {

		select {
		case <-local_done_ch:
			remaining -= 1
		case e := <-local_err_ch:
			err_ch <- e
			return
		default:
			// pass
		}
	}

}

func (p *RWProcessor) Name() string {
	return "s3"
}

func (p *RWProcessor) Flush() error {
	return nil
}
