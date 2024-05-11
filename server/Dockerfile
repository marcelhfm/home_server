FROM golang:1.22.2-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . . 

RUN CG0_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tcp_server .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/tcp_server .

EXPOSE 5000

CMD ["./tcp_server"]
