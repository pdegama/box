FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/box ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /
COPY --from=build /out/box /usr/local/bin/box
# config: mount box.yml at /etc/box.yml ; tls keys via a volume (e.g. /keys)
# MODE=server (default) or MODE=client
EXPOSE 25
CMD ["box"]
