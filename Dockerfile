# syntax=docker/dockerfile:1
FROM golang:1.24.3-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN GOWORK=off go build -o evosim

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/evosim ./evosim
EXPOSE 8080
CMD ["./evosim", "--web", "--web-port", "8080"]

