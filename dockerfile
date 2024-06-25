# syntax=docker/dockerfile:1
# Download Go Depedencies
FROM golang:1.22.0-alpine3.19 AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# Build local development image
FROM golang:1.22.0-alpine3.19 AS dev
COPY --from=base /go/bin /go/bin
COPY --from=base /go/pkg /go/pkg
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY . ./
EXPOSE 8888
# Build the Go Binary
FROM golang:1.22.0-alpine3.19 AS builder
COPY --from=base /go/bin /go/bin
COPY --from=base /go/pkg /go/pkg
WORKDIR /app
COPY . ./
EXPOSE 8888
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .
# Create Production Image
FROM gcr.io/distroless/base-debian11 AS release
COPY --from=builder /server /server
EXPOSE 8888
ENTRYPOINT ["/server"]
