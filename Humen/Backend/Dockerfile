# build
FROM golang:1.22-alpine AS builder
WORKDIR /app

# 先複製 mod 檔，拉依賴
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# 再複製 source code
COPY . .

# build app
RUN go build -o app .

# runtime
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/app /app/app
EXPOSE 8080
CMD ["./app"]
