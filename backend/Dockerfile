FROM golang:1.14.4-alpine3.12 AS builder

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/backend

FROM jrottenberg/ffmpeg:snapshot-alpine311
WORKDIR /bin/backend
COPY --from=builder /bin/backend /go/src/app/config/config.yaml /go/src/app/azure.auth ./
ENV AZURE_AUTH_LOCATION="/bin/backend/azure.auth"
EXPOSE 8081
ENTRYPOINT ["./backend"]