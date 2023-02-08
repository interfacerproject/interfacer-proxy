# SPDX-FileCopyrightText: 2023 Dyne.org foundation
#
# SPDX-License-Identifier: AGPL-3.0-or-later

curl -X POST localhost:8080/zenflows/api -d "query{instanceVariables{specs{specCurrency{id}specProjectDesign{id}specProjectProduct{id}specProjectService{id}}units{unitOne{id}}}}"
