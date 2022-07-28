# tlsclient


## **How To Use**
```go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	tlsclient "github.com/KakashiHatake324/tlsclient/v2"
	tls "github.com/KakashiHatake324/tlsclient/v2/utls"
)

func main() {

	jar, _ := cookiejar.New(nil)

	settings := CustomizedSettings{
		MaxHeaderListSize: 262144,

		// Set as true to include enable push in frames
		ServerPushSet: false,

		// Set as true to set enable push value to 1
		// or set as false to set enable push value to 0
		// if ServerPushSet is not true this will not get sent.
		ServerPushEnable: false,

		Priority: true,

		// Set value from 1 to 256
		PriorityWeight:       256,
		InitialWindowSize:    6291456,
		MaxConcurrentStreams: 1000,
		HeaderTableSize:      65536,
		WindowSizeIncrement:  15663105,
	}

	// proxy url is optional
	client, err := tlsclient.NewClient(tls.HelloChrome_100, jar, false, time.Duration(15), settings, "http://user:pass@ip:port")
	if err != nil {
		fmt.Println(err)
	}

	url := "https://www.google.com/"
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("accept-encoding", "gzip, deflate, br")
	req.Header.Add("accept-language", "en-GB,en;q=0.9")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("sec-ch-ua", "\".Not/A)Brand\";v=\"99\", \"Google Chrome\";v=\"103\", \"Chromium\";v=\"103\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")

	// Define headers order (optional)
	// if any header key is not in Header-Order that header will not get sent.
	req.Header.Add("header-order", "accept,accept-encoding,accept-language,cache-control,pragma,sec-ch-ua,sec-ch-ua-mobile,sec-ch-ua-platform,sec-fetch-dest,sec-fetch-mode,sec-fetch-site,sec-fetch-user,upgrade-insecure-requests,user-agent")

	// customize http/2 pseudo headers order (optional)
	// if chrome's user-agent is set then client will implicitly ovveride it
	// and set send pseudo headers in chrome's order
	// m = :method:, a = :authority:, s = :scheme:, p = :path:
	req.Header.Add("pseudo-headers-order", "s,m,p,a")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Brotli or Gzip Decompression
	returnedEncoding := res.Header.Get("Content-Encoding")
	if returnedEncoding == "gzip" {

		body = tlsclient.BodyDecompress(string(body), "gzip")

	} else if returnedEncoding == "br" {

		body = tleclient.BodyDecompress(string(body), "br")

	}

	fmt.Println(string(body))
}

```
