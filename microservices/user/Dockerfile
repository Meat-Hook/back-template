FROM golang:1.16-alpine AS builder

RUN apk update && apk add git && apk add ca-certificates

WORKDIR /build

COPY ./go.mod go.mod
COPY ./go.sum go.sum

RUN go mod download

COPY . .

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build ./microservices/user

FROM scratch

WORKDIR /bin

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/user user
COPY ./microservices/user/migrate/ migrate/

CMD ./user
