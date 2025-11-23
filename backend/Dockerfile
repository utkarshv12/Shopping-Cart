FROM golang:1.20-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /shopping-cart ./main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /shopping-cart /usr/local/bin/shopping-cart
EXPOSE 8080
ENV MYSQL_DSN=""
ENTRYPOINT ["/usr/local/bin/shopping-cart"]
