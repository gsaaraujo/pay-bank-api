FROM golang:1.25-alpine3.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o main ./cmd/main.go

FROM alpine:latest AS runtime
WORKDIR /app/
COPY --from=builder /app/main ./main
CMD ["./main"]