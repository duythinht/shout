FROM golang:1.20 as build
WORKDIR /opt/shout/src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /opt/shout/bin/fm ./cmd/fm/main.go
RUN go build -o /opt/shout/bin/music-station ./cmd/music-station/main.go

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates ffmpeg && rm -rf /var/lib/apt/lists/*
WORKDIR /opt/shout
COPY --from=build /opt/shout/bin/fm /usr/local/bin/
COPY --from=build /opt/shout/bin/music-station /usr/local/bin/
RUN mkdir -p /opt/shout/songs
CMD ["music-station"]