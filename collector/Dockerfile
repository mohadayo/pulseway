FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY *.go .
RUN go build -o collector .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/collector .
EXPOSE 8001
CMD ["./collector"]
