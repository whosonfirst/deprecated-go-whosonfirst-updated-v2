# https://blog.docker.com/2016/09/docker-golang/
# https://blog.golang.org/docker

# docker build -t wof-updated .

# build phase - see also:
# https://medium.com/travis-on-docker/multi-stage-docker-builds-for-creating-tiny-go-images-e0e1867efe5a
# https://medium.com/travis-on-docker/triple-stage-docker-builds-with-go-and-angular-1b7d2006cb88

FROM golang:alpine AS build-env

# https://github.com/gliderlabs/docker-alpine/issues/24

RUN apk add --update alpine-sdk

ADD . /go-whosonfirst-updated-v2

RUN cd /go-whosonfirst-updated-v2; make bin

FROM alpine

# RUN apk add --update bzip2 curl
# VOLUME /usr/local/data

WORKDIR /whosonfirst/bin/

COPY --from=build-env /go-whosonfirst-updated-v2/bin/wof-updated /whosonfirst/bin/wof-updated
COPY --from=build-env /go-whosonfirst-updated-v2/docker/entrypoint.sh /whosonfirst/bin/entrypoint.sh

EXPOSE 8080

ENTRYPOINT /whosonfirst/bin/entrypoint.sh
