# Builder image
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/github.com/atonem/miniflux-feed/
COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" -o /go/bin/miniflux-feed

# Main image
FROM scratch
# FROM golang:alpine

COPY --from=builder /go/bin/miniflux-feed /go/bin/miniflux-feed

# ENV MINIFLUX_URL
# ENV MINIFLUX_TOKEN
# ENV PORT
#
ENTRYPOINT ["/go/bin/miniflux-feed"]
