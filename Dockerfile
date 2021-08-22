FROM golang:1.16-alpine3.14 AS builder

RUN apk update && apk add --no-cache ca-certificates=20191127-r5

WORKDIR /build

COPY ./go.mod go.mod
COPY ./go.sum go.sum

RUN go mod download

COPY . .

ARG GO111MODULE=on
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

RUN go build .

FROM scratch

WORKDIR /bin

ARG SERVICE

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/back-template back-template
COPY ./cmd/user/migrate/ user/migrate/
COPY ./cmd/session/migrate/ session/migrate/
COPY ./cmd/file/migrate/ file/migrate/

ENTRYPOINT ["back-template"]
