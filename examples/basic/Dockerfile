# examples/basic/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o demo-app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /root/
COPY --from=builder /app/demo-app .
EXPOSE 8081
CMD ["./demo-app"] 