package dispatcher

import (
        "github.com/whosonfirst/go-webhookd"
        "github.com/whosonfirst/go-whosonfirst-readwrite/reader"
)

type ReadWriteDispatcher struct {
	webhookd.WebhookDispatcher
	reader 	reader.Reader
	// writer writer.Writer
}

func NewReadWriteDispatcher() (webhookd.WebhookDispatcher, error) {

     	d := ReadWriteDispatcher{

	}

	return &d, nil
}

func (d *ReadWriteDispatcher) Dispatch(body []byte) *webhookd.WebhookError {

     	
	return nil
}
