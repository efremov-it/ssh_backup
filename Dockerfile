FROM golang:1.23.0-alpine3.19 as go

WORKDIR /app
COPY . /app/

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=${VERSION}'" -o /bin/ssh-backup ./cmd

FROM alpine:3.19.2

WORKDIR /

COPY --from=go /bin/ssh-backup /bin/ssh-backup

