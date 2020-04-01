rtchat
===

This is a tiny and lightweight **experiment** to easily host live video conferences using the WebRTC protocol.

Since I love simple things, this tiny server is written in **Go**, final binary weights **~10mb**, uses **~8mb** of memory and provides:

- A STUN/TURN server using the [pion/turn](https://github.com/pion/turn) package that will forward streams if two peers cannot communicate directly,
- A websocket signaling server to create rooms and enables anyone in the same room to communicate using the WebRTC protocol,
- A tiny frontpage served by the server.

With the above stuff, you have everything you need to make peer to peer communications with WebRTC 💪

## Develop

Once you have the [Go](https://golang.org/dl/) language installed, just launch the server:

```console
$ go run cmd/server/main.go -debug
```

It will download dependencies and run the tiny server straigh away (see the list of available options below).

## Build

```console
$ go build -o rtchat cmd/server/main.go
```

## Deploy

This repository contains a `Dockerfile` to easily deploy this application (you must set environment variables `TURN_PORT` and `TURN_IP`). The final image weight about **~17mb** 😍

```console
$ docker build -t rtchat .
$ docker run -it --rm -p 5000:5000 -p 3478:3478/udp -e TURN_PORT=3478 -e TURN_IP=192.168.0.14 rtchat
```

## Usage

```console
Usage of rtchat:
  -debug
        Should we launch in the debug mode?
  -http-port int
        Web server listening port. (default 5000)
  -realm string
        Realm used by the turn server. (default "rtchat.io")
  -turn-ip string
        IP Address that TURN can be contacted on. Should be publicly available. (default "192.168.0.14")
  -turn-port int
        Listening port for the TURN/STUN endpoint. (default 3478)
```

If there is one parameter to keep in mind, it's the `-turn-ip` which represents the publicly available IP used by the TURN server to enables peer to communicate being NAT or proxys by forwarding all streams through the server.