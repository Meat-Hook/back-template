FROM golang:1.16-alpine AS builder

RUN apk update && apk add git && apk add ca-certificates

WORKDIR /build

COPY . .

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build ./internal/microservices/session

FROM scratch

WORKDIR /bin

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/session session
COPY ./internal/microservices/session/migrate migrate/

CMD ./session
