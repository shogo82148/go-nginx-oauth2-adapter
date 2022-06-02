FROM golang:1.18.3 as builder
WORKDIR /go/src/github.com/shogo82148/go-nginx-oauth2-adapter

# Faster builds in Docker with Go 1.11 or later
# https://container-solutions.com/faster-builds-in-docker-with-go-1-11/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -o go-nginx-oauth2-adapter ./cli/go-nginx-oauth2-adapter

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/shogo82148/go-nginx-oauth2-adapter/go-nginx-oauth2-adapter /usr/local/bin/go-nginx-oauth2-adapter
CMD ["/usr/local/bin/go-nginx-oauth2-adapter"]
