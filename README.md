<!--
SPDX-License-Identifier: AGPL-3.0-or-later
Copyright (C) 2022-2023 Dyne.org foundation <foundation@dyne.org>.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

# zenflows-proxy

A unifying HTTP proxy for services used by
[the interfacer-gui project](https://github.com/interfacerproject/interfacer-gui)


## Services

The front-end project,
[interfacer-gui](https://github.com/interfacerproject/interfacer-gui),
uses this project to unify interactions with the (micro) services
implemented for the Interfacer Project.

At the moment, the services implemented are:

* Zenflows - a [zenflows](https://github.com/interfacerproject/zenflows)
  instance.  It's the main back-end service.;
* Inbox - a [zenflows-inbox](https://github.com/interfacerproject/zenflows-inbox)
  instance.  It provides notifications and implements a subset of ActivityPub.
* Wallet - a [zenflows-wallet](https://github.com/interfacerproject/zenflows-wallet)
  instance.  It powers the economic model behind the Interfacer
  project.  It is an interface between interfacer-gui and planetmint.
* Location Autocomplition - provided by
  [Here API](https://autocomplete.search.hereapi.com/v1/autocomplete)
  to autocomplete location/place names on the map;
* Location Lookup - provided by
  [Here API](https://lookup.search.hereapi.com/v1/lookup) to lookup
  location/place names on the map;


## Bulding and executing

You may either choose to build from source, build and use the Docker
image, or just use the pre-built Docker image on
https://ghcr.io/interfacerproject/interfacer-gateway.  A set of
Ansible scripts are also provided, which could be used as well.

Building from source with Go might be a bit work due to the
dependencies.  Building from source with Docker or using Ansible
is recommended, as they will be easier to get started.

Each option is detailed in the following sections.


### Building from source with Go

Bulding with Go from the source requires Go version 1.18 or later.
If you have the Go toolchain and a POSIX-compliant make(1)
implementation installed (GNU make(1) works), you can just run:

	make

which builds a the service as the executable named `interfacer-gateway`.


### Building from source with Docker

Building with Docker requires nothing but the Docker tooling.  The
image produced will have all the dependencies needed to run this
service.

To build the image, you can run:

	docker build -t interfacer-gateway:latest .

which will name the image "interfacer-gateway".  Then, you can run:

	docker run --rm -p PORT:8080 \
		-e ZENFLOWS_URL=http://url-to-zenflows-instance \
    -e INTERFACER_DPP_URL=http://url-to-interfacer-dpp-instance \
		-e INBOX_URL=http://url-to-inbox-instance \
		-e WALLET_URL=http://url-to-wallet-instance \
		-e OSH_URL=http://url-to-osh-instance \
		-e HERE_KEY=api-key-for-here-dot-com \
		zenflows-osh

to start the service on port `PORT`.  The service will require you
to provide the environment variables listed above.  The variables
are descriped on the section "Configuration".


### Using the pre-built image

You may choose to just use the pre-built image, which is found at
https://ghcr.io/interfacerproject/interfacer-gateway.

To use that image, you can run:

	docker run --rm -p PORT:8080 \
		-e ZENFLOWS_URL=http://url-to-zenflows-instance \
    -e INTERFACER_DPP_URL=http://url-to-interfacer-dpp-instance \
		-e INBOX_URL=http://url-to-inbox-instance \
		-e WALLET_URL=http://url-to-wallet-instance \
		-e OSH_URL=http://url-to-osh-instance \
		-e HERE_KEY=api-key-for-here-dot-com \
		ghcr.io/interfacerproject/interfacer-gateway

which will start the service on port `PORT`. The service will require
you to provide the environment variables listed above.  The variables
are descriped on the section "Configuration".

You may optionally use a docker-compose.yml template like this as well:

```
version: "3.8"
services:
  gateway:
    container_name: gateway
    image: ghcr.io/interfacerproject/interfacer-gateway
    ports:
      # The service will be listening on port 3000 of the host
      # machine.
      - 3000:8080
    environment:
      ZENFLOWS_URL: http://url-to-zenflows-instance
      INTERFACER_DPP_URL: http://url-to-interfacer-dpp-instance
      INBOX_URL: http://url-to-inbox-instance
      WALLET_URL: http://url-to-wallet-instance
      OSH_URL: http://url-to-osh-instance
      HERE_KEY: api-key-for-here-dot-com
    stdin_open: true
    tty: true
```

As mentioned above, the environment variables are described on the
section "Configuration".

### Using the Ansible scripts

Deployment (most of the times) can be done using IaC with Ansible.
The deployment use docker-compose with a template that groups all the services required.
It can be started writing the `hosts.yaml` and running `make install` inside the `devops` directory.

In the `hosts.yaml` one has to provide:
 - `domain_name`: the FQDN of the server
 - `zenflows`: URL for the zenflows instanc3
 - `ifacer_log`: path to the log file
 - `port`: port for proxy on `localhost`
 - `here_api`: URL for the API of here.com
 - `inbox_port`: port for inbox on `localhost`
 - `wallet_port`: port for wallet on `localhost`
 - `zenflows_sk`: secret key for a user in zenflows
 - `zenflows_user`: username of the same user in zenflows

Warning: currently we are deploying in a LAN with private address, so you may need to modify `docker-compose.yaml.j2`


## Configuration

The configuration is done by providing environment variables.

Here is the list of them:

* `ZENFLOWS_URL` - the URL of
  [the zenflows instance](https://github.com/interfacerproject/zenflows)
  to which the requests will be proxied;
* `INTERFACER_DPP_URL` - the URL of
  [the interfacer-dpp instance](https://github.com/interfacerproject/interfacer-dpp)
  to which the requests will be proxied;
* `INBOX_URL` - the URL of
  [the zenflows-inbox instance](https://github.com/interfacerproject/zenflows-inbox)
  to which the requests should be proxied;
* `WALLET_URL` - the URL of
  [the zenflows-wallet instance](https://github.com/interfacerproject/zenflows-wallet)
  to which the requests should be proxied;
* `OSH_URL` - the URL of
  [the zenflows-osh instance](https://github.com/interfacerproject/zenflows-osh)
  to which the requests should be proxied;
* `HERE_KEY` - the API key required by the Here API.  More on that
  at: https://developer.here.com/documentation.


## Usage

The services are served over HTTP, and each one is accessed by
prefixing the path of the URL that is supposed to be sent to the
service with the identifier of the service.

For instance, let's assume that "/foo/and/whatnot" is supposed to be
sent to the "BAR" service, and the identifier of the service "BAR"
is "bar".  Then, the clients of the gateway sends the request to
the gateway with the url "/bar/foo/andwhatnot", assuming that
everything will be sent to the service "BAR".  As such, the provided
HTTP methods and headers will be forwarded to the service "BAR".


## Examples

See the subdirectory `examples/`.
