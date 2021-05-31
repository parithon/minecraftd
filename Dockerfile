FROM golang:latest as BUILD

WORKDIR /app

COPY . .

RUN go build .

FROM debian:stable-slim

RUN apt-get update && apt-get -y install --no-install-recommends ca-certificates curl 

WORKDIR /app

COPY --from=BUILD /app/minecraftd .

VOLUME [ "/data" ]

EXPOSE 19132/udp

ENTRYPOINT [ "/app/minecraftd" ]
