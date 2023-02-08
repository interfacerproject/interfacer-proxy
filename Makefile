# SPDX-FileCopyrightText: 2023 Dyne.org foundation
#
# SPDX-License-Identifier: AGPL-3.0-or-later

all:
	CGO_ENABLED=0 go build -o interfacer-gateway
