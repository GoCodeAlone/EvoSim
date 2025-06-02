# syntax=docker/dockerfile:1
FROM golang:1.24.3-alpine

# Install Air for live reloading
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

ENV PATH="/go/bin:${PATH}"

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
