FROM golang:1.24-alpine AS builder
WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest AS runtime
WORKDIR /app

RUN apk add --no-cache curl xz

RUN curl -fsSL https://github.com/typst/typst/releases/download/v0.12.0/typst-x86_64-unknown-linux-musl.tar.xz \
    | tar -xJ -C /usr/local/bin --strip-components=1 \
    && chmod +x /usr/local/bin/typst

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/pdf/templates ./templates

EXPOSE 8080 50051
COPY .env /app/.env
CMD ["./main"]