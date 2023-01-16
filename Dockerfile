FROM golang:1.19-alpine3.16 AS builder

ADD . /app
WORKDIR /app

RUN ENABLE_CGO=0 go build -o interfacer-gateway

FROM alpine:3.16 AS worker

ARG PORT=8080
ENV PORT=$PORT
ARG USER=app
ENV USER=$USER

ENV IFACER_LOG="/log"

WORKDIR /app

RUN addgroup -S "$USER" && adduser -SG "$USER" "$USER" && \
    install -d -m 0755 -o "$USER" -g "$USER" /log

COPY --from=builder /app/interfacer-gateway /app

USER $USER

EXPOSE $PORT

CMD ["/app/interfacer-gateway"]
