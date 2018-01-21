package copy

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
	"sync"
	"time"
)

// we pass along Copier, reader.Reader, writer.Writer separately largely because some tasks need to
// create discrete readers on a per instance basis (20180121/thisisaaronland)

func ReadWriteCopy(c Copier, r reader.Reader, w writer.Writer, t updated.Task, status_ch chan string, err_ch chan error) {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		msg := fmt.Sprintf("time to process %s task (%s) %v", c.Name(), t.String(), t2)
		status_ch <- msg
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clients := 10 // please put me somewhere common
	throttle_ch := make(chan bool, clients)

	for i := 0; i < clients; i++ {
		throttle_ch <- true
	}

	wg := new(sync.WaitGroup)

	count := len(t.Commits)

	for i, path := range t.Commits {

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

				status_ch <- fmt.Sprintf("[%s] %d/%d read %s", c.Name(), i, count, path)

				fh, err := r.Read(path)

				if err != nil {
					msg := fmt.Sprintf("[%s] %d/%d %s", c.Name(), i, count, err.Error())
					err_ch <- errors.New(msg)

					cancel()
					return
				}

				status_ch <- fmt.Sprintf("[%s] %d/%d write %s", c.Name(), i, count, path)

				err = w.Write(path, fh)

				fh.Close()

				if err != nil {
					msg := fmt.Sprintf("[%s] %d/%d %s", c.Name(), i, count, err.Error())
					err_ch <- errors.New(msg)

					cancel()
					return
				}
			}

		}(ctx, path, i)
	}

	wg.Wait()
}

type ReadWriteCopier struct {
	Copier
	reader reader.Reader
	writer writer.Writer // remember this might be a MultiWriter
}

func NewReadWriteCopier(r reader.Reader, w writer.Writer) (Copier, error) {

	pr := ReadWriteCopier{
		reader: r,
		writer: w,
	}

	return &pr, nil
}

func (pr *ReadWriteCopier) ProcessTask(t updated.Task, status_ch chan string, err_ch chan error, done_ch chan bool) {

	defer func() {
		done_ch <- true
	}()

	ReadWriteCopy(pr, pr.reader, pr.writer, t, status_ch, err_ch)
}

func (p *ReadWriteCopier) Name() string {
	return "readwrite-copy"
}

func (p *ReadWriteCopier) Flush() error {
	return nil
}
