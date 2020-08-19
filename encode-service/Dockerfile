FROM golang:1.14.4-alpine3.12 AS builder

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/encode-service

FROM jrottenberg/ffmpeg:snapshot-alpine311
WORKDIR /bin/encode-service
COPY --from=builder /bin/encode-service /go/src/app/encode/. ./
EXPOSE 80
ENTRYPOINT ["./encode-service", "-f", "/downloads"]