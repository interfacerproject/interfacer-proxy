package main

import (
	"errors"
	"fmt"
	"github.com/interfacerproject/interfacer-gateway/config"
	"github.com/interfacerproject/interfacer-gateway/logger"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

var conf *config.Config

const clientTimeout = 10 * time.Second

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
		}).Error("client: could not create request: %s", err.Error())

		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "client: could not create request\n")
		return
	}
	req.Header = r.Header
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer func() {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}()
	}
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"app":   p.name,
			"host":  r.RemoteAddr,
			"error": err.Error(),
		}).Error("client: error making http request: %s", err.Error())

		w.Header().Add("access-control-allow-origin", "*")
		w.Header().Add("access-control-allow-credentials", "false")
		w.Header().Add("access-control-allow-methods", "POST, GET, DELETE, PUT, OPTIONS, PATCH")
		w.Header().Add("access-control-allow-headers", "*")

		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "client: error making http request to %s\n", p.name)
		return
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
