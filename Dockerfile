# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/misterhex/gogogocrawler
WORKDIR /go/src/github.com/misterhex/gogogocrawler
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get .
RUN go build /go/src/github.com/misterhex/gogogocrawler/main.go

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/src/github.com/misterhex/gogogocrawler/main


