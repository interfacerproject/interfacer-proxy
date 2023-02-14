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

package main

import (
	"errors"
	"fmt"
	"github.com/interfacerproject/interfacer-gateway/config"
	"github.com/interfacerproject/interfacer-gateway/logger"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var conf *config.Config

const clientTimeout = 10 * time.Second

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}
var transport = &http.Transport{
	DisableKeepAlives:     true,
	Proxy:                 http.ProxyFromEnvironment,
	DialContext:           dialer.DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// from https://pkg.go.dev/net/http#pkg-overview
// Clients and Transports are safe for concurrent use by multiple goroutines
// and for efficiency should only be created once and re-used.
// TODO: Look at https://mauricio.github.io/golang-proxies
var client = &http.Client{
	Timeout: clientTimeout,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// A ProxiedHost reppresent an host we will redirect to
// the `name` is just a string identifier (this will be the first name in the request path)
// `buildUrl` takes the input and create the url that will be used in the request of the proxy
// TODO: implement an authentication mechanism
type ProxiedHost struct {
	name string
	// authenticatd bool
	buildUrl func(*url.URL) *url.URL
}

// Currently know host we will proxy to
var proxiedHosts = []ProxiedHost{
	ProxiedHost{
		name: "zenflows",
		buildUrl: func(u *url.URL) *url.URL {
			return conf.ZenflowsURL.JoinPath(u.EscapedPath()[len("/zenflows"):])
		},
	},
	ProxiedHost{
		name: "location-autocomplete",
		buildUrl: func(u *url.URL) *url.URL {
			values := u.Query()
			values.Add("apiKey", conf.HereKey)
			currentUrl, _ := url.Parse("https://autocomplete.search.hereapi.com/v1/autocomplete")
			currentUrl.RawQuery = values.Encode()
			return currentUrl
		},
	},
	ProxiedHost{
		name: "location-lookup",
		buildUrl: func(u *url.URL) *url.URL {
			values := u.Query()
			values.Add("apiKey", conf.HereKey)
			currentUrl, _ := url.Parse("https://lookup.search.hereapi.com/v1/lookup")
			currentUrl.RawQuery = values.Encode()
			return currentUrl
		},
	},
	ProxiedHost{
		name: "inbox",
		buildUrl: func(u *url.URL) *url.URL {
			newurl := conf.InboxURL.JoinPath(u.EscapedPath()[len("/inbox"):])
			newurl.RawQuery = u.RawQuery
			return newurl
		},
	},
	ProxiedHost{
		name: "wallet",
		buildUrl: func(u *url.URL) *url.URL {
			newurl := conf.WalletURL.JoinPath(u.EscapedPath()[len("/wallet"):])
			newurl.RawQuery = u.RawQuery
			return newurl
		},
	},
	ProxiedHost{
		name: "osh",
		buildUrl: func(u *url.URL) *url.URL {
			return conf.OSHURL.JoinPath(u.EscapedPath()[len("/osh"):])
		},
	},
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	w.Header().Add("access-control-allow-credentials", "false")
	w.Header().Add("access-control-allow-methods", "POST, GET, DELETE, PUT, OPTIONS, PATCH")
	w.Header().Add("access-control-allow-headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	for _, host := range proxiedHosts {
		fmt.Fprintf(w, "/%s/\n", host.name)
	}
}

func (p *ProxiedHost) proxyRequest(w http.ResponseWriter, r *http.Request) {
	reqUrl := p.buildUrl(r.URL).String()
	req, err := http.NewRequest(r.Method, reqUrl, r.Body)
	// Can't really fail due to method, url, and the body are provided by the std lib.
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"app":   p.name,
			"host":  r.RemoteAddr,
			"error": err.Error(),
		}).Errorf("client: could not create request: %s", err.Error())

		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "client: could not create request\n")
		return
	}
	req.Header = r.Header
	maxRetry := 3
	var res *http.Response = nil
	for i := 0; ; i = i + 1 {
		var err error
		res, err = client.Do(req)
		if err == nil {
			break
		}
		if err != nil && i == maxRetry {
			logger.Log.WithFields(logrus.Fields{
				"app":   p.name,
				"host":  r.RemoteAddr,
				"error": err.Error(),
			}).Errorf("client: error making http request: %s", err.Error())

			w.Header().Add("access-control-allow-origin", "*")
			w.Header().Add("access-control-allow-credentials", "false")
			w.Header().Add("access-control-allow-methods", "POST, GET, DELETE, PUT, OPTIONS, PATCH")
			w.Header().Add("access-control-allow-headers", "*")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			fmt.Fprintf(w, "client: error making http request to %s\n", p.name)

			return
		}
	}
	if res != nil && res.Body != nil {
		defer func() {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}()
	}
	// Read all the headers
	for name, headers := range res.Header {
		// Iterate all headers with one name (e.g. Content-Type)
		for _, hdr := range headers {
			w.Header().Add(name, hdr)
		}
	}
	logger.Log.WithFields(logrus.Fields{
		"app":  p.name,
		"url":  reqUrl,
		"host": r.RemoteAddr,
	}).Info("Proxy request")
	io.Copy(w, res.Body)
}

func (p *ProxiedHost) addHandle() (string, func(w http.ResponseWriter, r *http.Request)) {
	return "/" + p.name + "/", p.proxyRequest
}

func main() {
	logger.InitLog(os.Getenv("IFACER_LOG"))

	var err error
	conf, err = config.NewEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configs couldn't be loaded: %s\n", err.Error())
		os.Exit(1)
	}

	http.HandleFunc("/", getRoot)
	for i := 0; i < len(proxiedHosts); i++ {
		http.HandleFunc(proxiedHosts[i].addHandle())
	}

	fmt.Fprintf(os.Stderr, "starting server on %s\n", conf.Addr)
	err = http.ListenAndServe(conf.Addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(os.Stderr, "server closed")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %s\n", err.Error())
		os.Exit(2)
	}
}
