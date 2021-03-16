FROM golang:1.16-alpine AS builder

RUN apk update && apk add git && apk add ca-certificates

WORKDIR /build

COPY ./go.mod go.mod
COPY ./go.sum go.sum

RUN go mod download

COPY . .

ARG GO111MODULE=on
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG SERVICE

RUN go build ./microservices/$SERVICE

FROM scratch

WORKDIR /bin

ARG SERVICE

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/$SERVICE $SERVICE
COPY ./microservices/$SERVICE/migrate/ migrate/

CMD ./$SERVICE
