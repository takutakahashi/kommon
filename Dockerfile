FROM golang:1.23-alpine AS builder

# Docker API バージョンを明示的に設定
ENV DOCKER_API_VERSION=1.45

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o kommon

FROM alpine:latest

# Docker API バージョンを明示的に設定
ENV DOCKER_API_VERSION=1.45

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/kommon .

ENTRYPOINT ["./kommon"]