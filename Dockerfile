FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy Go files and go.mod
COPY go.mod .
COPY main.go .
COPY exhibits.json .
COPY qm_data.json .

# Build the Go application
# Disable CGO and build statically to avoid runtime issues in Docker
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -ldflags="-w -s" -o stub-exhibits-api main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates netcat-openbsd

WORKDIR /app

# Copy binary and data from builder
COPY --from=builder /app/stub-exhibits-api .
COPY --from=builder /app/exhibits.json .
COPY --from=builder /app/qm_data.json .

EXPOSE 8000

CMD ["./stub-exhibits-api"]

