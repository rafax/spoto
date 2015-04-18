FROM golang

ADD *.go /go/src/github.com/rafax/spoto/

RUN go get github.com/rafax/spoto
RUN go install github.com/rafax/spoto

ENTRYPOINT /go/bin/spoto

EXPOSE 3000