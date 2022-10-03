package main

import (
    "strings"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
    "io/ioutil"
    "net/url"
)

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
var proxiedHosts = []ProxiedHost {
    ProxiedHost {
        name: "zenflows",
        buildUrl: func(u *url.URL) *url.URL {
            paths := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
            currentUrl, _ := url.Parse("https://fcos.interfacer.dyne.org/api")
            currentUrl.Path = paths[1]
            return currentUrl
        },
    },
    ProxiedHost {
        name: "location-autocomplete",
        buildUrl: func(u *url.URL) *url.URL {
            values := u.Query()
            values.Add("apiKey", os.Getenv("HERE_API"))
            currentUrl, _ := url.Parse("https://autocomplete.search.hereapi.com/v1/autocomplete")
            currentUrl.RawQuery = values.Encode()
            return currentUrl
        },
    },
    ProxiedHost {
        name: "location-lookup",
        buildUrl: func(u *url.URL) *url.URL {
            values := u.Query()
            values.Add("apiKey", os.Getenv("HERE_API"))
            currentUrl, _ := url.Parse("https://lookup.search.hereapi.com/v1/lookup")
            currentUrl.RawQuery = values.Encode()
            return currentUrl
        },
    },
}

func getRoot(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "Here I'll put the list of proxied host\n")
}

func (p *ProxiedHost) proxyRequest (w http.ResponseWriter, r *http.Request) {
    // The request has the path /destination_name/desired_path
    if !strings.HasPrefix(r.URL.Path, "/" + p.name + "/") {
        fmt.Printf("wrong destination")
        os.Exit(1)
    }

    req, err := http.NewRequest(r.Method, p.buildUrl(r.URL).String(), r.Body)
    if err != nil {
        msg := fmt.Sprintf("client: could not create request: %s\n", err)
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte(msg))
        return
    }
    req.Header = r.Header
    res, err := client.Do(req)
    if res != nil && res.Body != nil {
        defer func() {
            io.Copy(ioutil.Discard, res.Body)
            res.Body.Close()
        }()
    }
    if err != nil {
        msg := fmt.Sprintf("client: error making http request: %s\n", err)
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte(msg))
        return
    }
    // Read all the headers
    for name, headers := range res.Header {
        // Iterate all headers with one name (e.g. Content-Type)
        for _, hdr := range headers {
            w.Header().Add(name, hdr)
        }
    }
    io.Copy(w, res.Body)
}

func (p *ProxiedHost) addHandle() (string, func(w http.ResponseWriter, r *http.Request)) {
    return "/" + p.name + "/", p.proxyRequest
}

func main() {
    http.HandleFunc("/", getRoot)
    for i := 0; i<len(proxiedHosts); i++{
        http.HandleFunc(proxiedHosts[i].addHandle())
    }

    portStr := fmt.Sprintf(":%s", os.Getenv("PORT"))

    fmt.Printf("Starting server on port %s\n", os.Getenv("PORT"))
    err := http.ListenAndServe(portStr, nil)
    if errors.Is(err, http.ErrServerClosed) {
        fmt.Printf("server closed\n")
    } else if err != nil {
        fmt.Printf("error starting server: %s\n", err)
        os.Exit(1)
    }
}
