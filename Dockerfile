FROM golang:1.18-alpine AS builder
WORKDIR /usr/local/go/src/webdav-sync
COPY . /usr/local/go/src/webdav-sync
RUN go env -w GO111MODULE="on"
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN GOOS=linux GOARCH=amd64 go build main.go

 
 
FROM alpine AS runner
WORKDIR /usr/local/go/src/webdav-sync
RUN apk add tzdata
RUN apk add bash
RUN apk add wget
COPY --from=builder /usr/local/go/src/webdav-sync/main .
ENTRYPOINT ["./main"]