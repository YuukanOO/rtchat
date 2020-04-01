FROM golang:1.14-alpine AS builder
WORKDIR /go/src/app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o rtchat cmd/server/main.go

FROM alpine
COPY --from=builder /go/src/app/rtchat .
COPY --from=builder /go/src/app/static ./static
COPY --from=builder /go/src/app/templates ./templates
EXPOSE 5000 3478/udp
CMD ["sh", "-c", "./rtchat -realm=$REALM -turn-port=$TURN_PORT -turn-ip=$TURN_IP"]
