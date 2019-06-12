package cmd

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
	"log"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"net"

	rp "github.com/zckevin/reverseproxy/lib"
)

func NewReverseProxy(bindAddr, listenPort, socks5Addr, payloadPath string) (*rp.ReverseProxy, error) {
	var err error
	remote, _ := url.Parse("https://bt.byr.cn")

	// TODO: not working when local server works well but remote conn times out.
	defaultDialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	dialFunc := defaultDialer.Dial
	 
	if len(socks5Addr) > 0 {
		dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, defaultDialer)
		if err != nil {
			log.Println("Dial socks5 proxy error: ", err)
			return nil, err
		}
		dialFunc = dialer.Dial
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial:            dialFunc,
	}
	err = http2.ConfigureTransport(tr)
	if err != nil {
		log.Println("Configure http/2 error: ", err)
		return nil, err
	}

	marker := regexp.MustCompile(`<head>`)
	payload := []byte(fmt.Sprintf("<script src='%s'></script>", payloadPath))

	return &rp.ReverseProxy{
		Director: func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Host = remote.Host
			req.Header.Set("X-Origin-Host", remote.Host)
			// req.Header.Del("Host")
			// req.Header.Set("Host", remote.Host)
			req.URL.Scheme = "https"
			req.URL.Host = remote.Host
			log.Println(">>>", req)
		},
		Transport: tr,
		ModifyResponse: func(resp *http.Response) error {
			if resp.StatusCode == 302 && len(resp.Header.Get("Location")) > 0 {
				u, err := url.Parse(resp.Header.Get("Location"))
				if err != nil {
					return err
				}
				u.Host = bindAddr + ":" + listenPort
				u.Scheme = "http"
				resp.Header.Set("Location", u.String())
			}

			if resp.Header.Get("Content-Encoding") == "gzip" &&
				strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
				resp.Header.Del("Content-Encoding")
				resp.Header.Add("_gziped_html", "1")
			}

			log.Println("<<<", resp)
			return nil
		},
		Injector: &rp.CodeInjector{
			payload,
			marker,
		},
	}, nil
}
