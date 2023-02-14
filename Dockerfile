# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2022-2023 Dyne.org foundation <foundation@dyne.org>.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

ARG GOVER=1.18
FROM golang:$GOVER-bullseye AS builder
ENV GONOPROXY=

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN ENABLE_CGO=0 go build -o interfacer-proxy .


FROM dyne/devuan:chimaera

ARG PORT=8080
ENV ADDR=:$PORT
ARG USER=app

ENV IFACER_LOG="/log"

RUN addgroup --system "$USER" && adduser --system --ingroup "$USER" "$USER" && \
      install -d -m 0755 -o "$USER" -g "$USER" /log

WORKDIR /app

COPY --from=builder /app/interfacer-proxy .

USER $USER

EXPOSE $PORT

CMD ["./interfacer-proxy"]
