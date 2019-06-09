package main

import (
    "log"
    "net/url"
    "net/http"
    "strings"
    "regexp"
    "crypto/tls"
    "golang.org/x/net/proxy"
    "golang.org/x/net/http2"

    rp "reverseproxy"
)

func NewReverseProxy() *rp.ReverseProxy {
    remote, err := url.Parse("https://bt.byr.cn")
    if err != nil {
        panic(err)
    }

    dialer, err := proxy.SOCKS5("tcp", "localhost:9090", nil, nil)
    if err != nil {
        panic(err)
    }

    tr := &http.Transport {
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true,},
        Dial: dialer.Dial,
    }
    err = http2.ConfigureTransport(tr)
    log.Println(err)

    marker := regexp.MustCompile(`<head>`)
    payload := []byte("<script src='__bt/inject.js'></script>")

    return &rp.ReverseProxy {
        Director: func(req *http.Request) {
            req.Header.Add("X-Forwarded-Host", req.Host)
                    req.Host = remote.Host
                    req.Header.Set("X-Origin-Host", remote.Host)
                    // req.Header.Del("Host")
                    // req.Header.Set("Host", remote.Host)
                            req.URL.Scheme = "https"
                                    req.URL.Host = remote.Host
            // log.Println(req)
        },
        Transport: tr,
        ModifyResponse: func(resp *http.Response) error {
            if resp.StatusCode == 302 && len(resp.Header.Get("Location")) > 0 {
                u, err := url.Parse(resp.Header.Get("Location"))
                if err != nil {
                    return err
                }
                u.Host = "localhost:9988"
                u.Scheme = "http"
                resp.Header.Set("Location", u.String())
            }

            if resp.Header.Get("Content-Encoding") == "gzip" &&
                strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
                resp.Header.Del("Content-Encoding")
                resp.Header.Add("_gziped_html", "1")
            }

            log.Println(resp.Header)
            return nil
        },
        Injector: &rp.CodeInjector {
            payload,
            marker,
        },
    }
}

func main() {
    rp := NewReverseProxy()
    log.Fatal(http.ListenAndServe(":9988", rp))
}
