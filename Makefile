CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-updated-v2; then rm -rf src/github.com/whosonfirst/go-whosonfirst-updated-v2; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-updated-v2
	cp -r processor src/github.com/whosonfirst/go-whosonfirst-updated-v2/
	cp *.go src/github.com/whosonfirst/go-whosonfirst-updated-v2/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-csv"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-readwrite/..."
	@GOPATH=$(GOPATH) go get -u "gopkg.in/redis.v1"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt processor/*.go
	go fmt *.go

bin: 	rmdeps self
	@GOPATH=$(shell pwd) go build -o bin/wof-updated cmd/wof-updated.go
