package main

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
    "io/ioutil"
)

func httpClient() *http.Client {
    clientTimeout := 10 * time.Second

    client := &http.Client{
        Timeout: clientTimeout,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    return client
}

func getRoot(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("got / request\n")
    io.WriteString(w, "This is my website!\n")
}

// TODO: Look at https://mauricio.github.io/golang-proxies
func makeProxy(prefix string, proxiedHost string) func(http.ResponseWriter, *http.Request){
    // For there is a client for each host I want to proxy
    client := httpClient()

    return func (w http.ResponseWriter, r *http.Request) {
        url := fmt.Sprintf("%s/%s", proxiedHost, r.URL.Path[len(prefix)+1:])

        req, err := http.NewRequest(r.Method, url, r.Body)
        if err != nil {
            fmt.Printf("client: could not create request: %s\n", err)
            os.Exit(1)
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
            fmt.Printf("client: error making http request: %s\n", err)
            os.Exit(1)
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
}


func main() {
    //http.HandleFunc("/", getRoot)
    http.HandleFunc("/zenflows/", makeProxy("/zenflows", "https://fcos.interfacer.dyne.org"))
    err := http.ListenAndServe(":8080", nil)
    if errors.Is(err, http.ErrServerClosed) {
        fmt.Printf("server closed\n")
    } else if err != nil {
        fmt.Printf("error starting server: %s\n", err)
        os.Exit(1)
    }
}
