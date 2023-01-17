package tlsclient

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/net/proxy"

	utls "github.com/KakashiHatake324/tlsclient/v2/utls"
)

func NewClient(clientHello utls.ClientHelloID, jar http.CookieJar, redirect bool, timeout time.Duration, settings CustomizedSettings, host string, cert string, proxyUrl ...string) (http.Client, error) {

	var client http.Client
	var newerror error

	if cert != "" {
		certMutex.Lock()
		if _, ok := loadedCerts[host]; !ok {
			loadedCerts[host] = cert
		}
		certMutex.Unlock()
	}

	log.Println(host)

	if redirect {
		if len(proxyUrl) > 0 && len(proxyUrl[0]) > 0 {
			dialer, err := newConnectDialer(proxyUrl[0])
			if err != nil {
				client, newerror = http.Client{}, err
			}
			client = http.Client{
				Transport: newRoundTripper(clientHello, settings, dialer),
				Jar:       jar,
				Timeout:   timeout * time.Second,
			}
		} else {
			client = http.Client{
				Transport: newRoundTripper(clientHello, settings, proxy.Direct),
				Jar:       jar,
				Timeout:   timeout * time.Second,
			}
		}

	} else {
		if len(proxyUrl) > 0 && len(proxyUrl[0]) > 0 {
			dialer, err := newConnectDialer(proxyUrl[0])
			if err != nil {
				client, newerror = http.Client{}, err
			}
			client = http.Client{
				Transport:     newRoundTripper(clientHello, settings, dialer),
				Jar:           jar,
				Timeout:       timeout * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
			}
		} else {
			client = http.Client{
				Transport:     newRoundTripper(clientHello, settings, proxy.Direct),
				Jar:           jar,
				Timeout:       timeout * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
			}
		}
	}

	return client, newerror
}
