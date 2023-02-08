# SPDX-FileCopyrightText: 2023 Dyne.org foundation
#
# SPDX-License-Identifier: AGPL-3.0-or-later

FROM golang:1.19-bullseye AS builder

ADD . /app
WORKDIR /app

RUN ENABLE_CGO=0 go build -o interfacer-gateway

FROM dyne/devuan:chimaera AS worker

ARG PORT=8080
ENV PORT=$PORT
ARG USER=app
ENV USER=$USER

ENV IFACER_LOG="/log"

WORKDIR /app

RUN addgroup --system "$USER" && adduser --system --ingroup "$USER" "$USER" && \
    install -d -m 0755 -o "$USER" -g "$USER" /log

COPY --from=builder /app/interfacer-gateway /app

USER $USER

EXPOSE $PORT

CMD ["/app/interfacer-gateway"]
