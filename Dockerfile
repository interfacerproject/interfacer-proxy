# SPDX-FileCopyrightText: 2023 Dyne.org foundation
#
# SPDX-License-Identifier: AGPL-3.0-or-later

FROM golang:1.19-alpine3.17 AS builder

WORKDIR /app
ADD . .

RUN ENABLE_CGO=0 go build -o interfacer-gateway .

FROM alpine:3.17

ARG USER=zenflows GROUP=zenflows
RUN addgroup -S "$GROUP" && adduser -SG"$GROUP" "$USER"

ENV IFACER_LOG="/log"
RUN install -d -m 0755 -o "$USER" -g "$GROUP" "$IFACER_LOG"

USER "$USER"

ARG PORT=8080
ENV PORT=$PORT
EXPOSE $PORT

WORKDIR /app

COPY --from=builder --chown="$USER:$GROUP" /app/interfacer-gateway .

CMD ["./interfacer-gateway"]
