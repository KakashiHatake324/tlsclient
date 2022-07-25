# tlsclient

init 

```go
jar, _ := cookiejar.New(nil)
client, err = tlsclient.NewClient(tls.HelloChrome_100, jar, false, time.Duration(15), "http://PROXYUSER@PROXYIP")
if err != nil {
	return err
}
```
