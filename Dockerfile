# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

ADD *.go /go/src/github.com/rafax/spoto/

RUN go get github.com/rafax/spoto
RUN go install github.com/rafax/spoto

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/spoto

# Document that the service listens on port 8080.
EXPOSE 3000