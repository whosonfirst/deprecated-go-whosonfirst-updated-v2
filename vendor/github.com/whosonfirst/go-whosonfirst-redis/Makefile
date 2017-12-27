CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test ! -d src; then mkdir src; fi
	if test ! -d src/github.com/whosonfirst/go-whosonfirst-redis/pubsub; then mkdir -p src/github.com/whosonfirst/go-whosonfirst-redis/pubsub; fi
	if test ! -d src/github.com/whosonfirst/go-whosonfirst-redis/resp; then mkdir -p src/github.com/whosonfirst/go-whosonfirst-redis/resp; fi
	cp pubsub/*.go src/github.com/whosonfirst/go-whosonfirst-redis/pubsub/
	cp resp/*.go src/github.com/whosonfirst/go-whosonfirst-redis/resp/
	cp -r vendor/src/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "gopkg.in/redis.v1"

vendor-deps: deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/publish cmd/publish.go
	@GOPATH=$(GOPATH) go build -o bin/subscribe cmd/subscribe.go
	@GOPATH=$(GOPATH) go build -o bin/pubsubd cmd/pubsubd.go

fmt:
	go fmt cmd/*.go
	go fmt pubsub/*.go
	go fmt resp/*.go

pub:
	./bin/publish -redis-channel debug -pubsubd -debug -

sub:
	./bin/subscribe -redis-channel debug
