# stage 1: builder
FROM golang:1.16-alpine as builder

RUN apk add --no-cache git curl
COPY . /usr/src/Burrow
WORKDIR /usr/src/Burrow

RUN go mod tidy && go build -o /tmp/ ./...

# stage 2: runner
FROM alpine:3.13

COPY --from=builder /tmp/burrow /app/
COPY ./entrypoint.sh /etc/burrow/entrypoint.sh

CMD ["/etc/burrow/entrypoint.sh"]
