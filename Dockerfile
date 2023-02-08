# SPDX-FileCopyrightText: 2023 Dyne.org foundation
#
# SPDX-License-Identifier: AGPL-3.0-or-later

FROM golang:1.19-bullseye AS builder

WORKDIR /app
COPY . .

RUN ENABLE_CGO=0 go build -o interfacer-proxy .

FROM dyne/devuan:chimaera AS worker

ARG PORT=8080
ENV ADDR=:$PORT
ARG USER=app
ENV USER=$USER

ENV IFACER_LOG="/log"

RUN addgroup --system "$USER" && adduser --system --ingroup "$USER" "$USER" && \
      install -d -m 0755 -o "$USER" -g "$USER" /log

WORKDIR /app

COPY --from=builder /app/interfacer-proxy .

USER $USER

EXPOSE $PORT

CMD ["./interfacer-proxy"]
