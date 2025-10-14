// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2022-2023 Dyne.org foundation <foundation@dyne.org>.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
)

type Config struct {
	// Addr is the address string to bind and listen on.
	// It will always be formed right, since it is put
	// together with net.JoinHostPort().
	Addr string

	// ZenflowsURL is the URL of the zenflows instance to which
	// we proxy request.  More on that at the project page:
	// https://github.com/interfacerproject/zenflows.
	ZenflowsURL *url.URL

	// InboxURL is the URL of the zenflows-inbox instance to
	// which we proxy request.  More on that at the project page:
	// https://github.com/interfacerproject/zenflows-inbox.
	InboxURL *url.URL

	// WalletURL is the URL of the zenflows-wallet instance to
	// which we proxy request.  More on that at the project page:
	// https://github.com/interfacerproject/zenflows-wallet
	WalletURL *url.URL

	// InterfacerDPPURL is the URL of the interfacer-dpp instance to
	// which we proxy request.  More on that at the project page:
	// https://github.com/interfacerproject/interfacer-dpp
	InterfacerDPPURL *url.URL

	// OSHURL is the URL of the zenflows-osh instance to which
	// we proxy request.  More on that at the project page:
	// https://github.com/interfacerproject/zenflows-osh
	OSHURL *url.URL

	// The opaque API key of here.com.  More on that at:
	// https://developer.here.com/documentation
	HereKey string
}

// NewEnv() fetches configuration options from the environment.  If
// a required option is not available or is malformed, it will error
// out.
func NewEnv() (*Config, error) {
	c := &Config{}

	s, ok := os.LookupEnv("ADDR")
	if !ok {
		return nil, errors.New(`"ADDR" must be provided`)
	}
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil, fmt.Errorf(`"ADDR" is malformed: %w`, err)
	}
	c.Addr = net.JoinHostPort(host, port)

	u, err := fetchURL("ZENFLOWS_URL")
	if err != nil {
		return nil, err
	}
	c.ZenflowsURL = u

	u, err = fetchURL("INBOX_URL")
	if err != nil {
		return nil, err
	}
	c.InboxURL = u

	u, err = fetchURL("WALLET_URL")
	if err != nil {
		return nil, err
	}
	c.WalletURL = u

	u, err = fetchURL("INTERFACER_DPP_URL")
	if err != nil {
		return nil, err
	}
	c.InterfacerDPPURL = u

	u, err = fetchURL("OSH_URL")
	if err != nil {
		return nil, err
	}
	c.OSHURL = u

	s = os.Getenv("HERE_KEY")
	if s == "" {
		return nil, errors.New(`"HERE_KEY" must be provided`)
	}
	c.HereKey = s

	return c, nil
}

func fetchURL(env string) (*url.URL, error) {
	s, ok := os.LookupEnv(env)
	if !ok {
		return nil, fmt.Errorf("%q must be provided", env)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("%q is malformed: %w", env, err)
	}

	// not a url, possibly a url
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("%q is malformed: not a url", env)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("%q is malformed: invalid scheme; must be http(s)", env)
	}

	// normalize it: take only what we need
	u = &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "/",
	}

	return u, nil
}
