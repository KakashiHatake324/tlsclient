package tlsclient

import (
	"net/http"
	"time"

	"golang.org/x/net/proxy"

	utls "github.com/KakashiHatake324/tlsclient/v2/utls"
)

func NewClient(clientHello utls.ClientHelloID, jar http.CookieJar, redirect bool, timeout time.Duration, settings CustomizedSettings, host string, cert string, proxyUrl ...string) (*http.Client, error) {

	var client *http.Client
	var newerror error

	if cert != "" {
		certMutex.Lock()
		if _, ok := loadedCerts[host]; !ok {
			loadedCerts[host] = cert
		}
		certMutex.Unlock()
	}

	if redirect {
		if len(proxyUrl) > 0 && len(proxyUrl[0]) > 0 {
			dialer, err := newConnectDialer(proxyUrl[0])
			if err != nil {
				client, newerror = nil, err
			}
			client = &http.Client{
				Transport: newRoundTripper(clientHello, settings, dialer),
				Jar:       jar,
				Timeout:   timeout,
			}
		} else {
			client = &http.Client{
				Transport: newRoundTripper(clientHello, settings, proxy.Direct),
				Jar:       jar,
				Timeout:   timeout,
			}
		}

	} else {
		if len(proxyUrl) > 0 && len(proxyUrl[0]) > 0 {
			dialer, err := newConnectDialer(proxyUrl[0])
			if err != nil {
				client, newerror = nil, err
			}
			client = &http.Client{
				Transport:     newRoundTripper(clientHello, settings, dialer),
				Jar:           jar,
				Timeout:       timeout,
				CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
			}
		} else {
			client = &http.Client{
				Transport:     newRoundTripper(clientHello, settings, proxy.Direct),
				Jar:           jar,
				Timeout:       timeout,
				CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
			}
		}
	}

	return client, newerror
}
