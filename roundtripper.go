package tlsclient

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/KageSolutions/tlsclient/net/http2"
	"golang.org/x/net/proxy"

	utls "gitlab.com/yawning/utls.git"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type roundTripper struct {
	sync.Mutex

	clientHelloId utls.ClientHelloID

	cachedConnections map[string]net.Conn
	cachedTransports  map[string]http.RoundTripper

	dialer       proxy.ContextDialer
	originalHost string
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	addr := rt.getDialTLSAddr(req)
	if _, ok := rt.cachedTransports[addr]; !ok {
		if err := rt.getTransport(req, addr); err != nil {
			return nil, err
		}
	}
	return rt.cachedTransports[addr].RoundTrip(req)
}

func (rt *roundTripper) getTransport(req *http.Request, addr string) error {
	switch strings.ToLower(req.URL.Scheme) {
	case "http":
		rt.cachedTransports[addr] = &http.Transport{DialContext: rt.dialer.DialContext}
		return nil
	case "https":
	default:
		return fmt.Errorf("invalid URL scheme: [%v]", req.URL.Scheme)
	}

	_, err := rt.dialTLS(context.Background(), "tcp", addr)
	switch err {
	case errProtocolNegotiated:
	case nil:
		// Should never happen.
		log.Println("dialTLS returned no error when determining cachedTransports")
	default:
		return err
	}

	return nil
}

func (rt *roundTripper) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	rt.Lock()
	defer rt.Unlock()

	var host string

	// If we have the connection from when we determined the HTTPS
	// cachedTransports to use, return that.
	if conn := rt.cachedConnections[addr]; conn != nil {
		delete(rt.cachedConnections, addr)
		return conn, nil
	}

	rawConn, err := rt.dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	if host, _, err = net.SplitHostPort(addr); err != nil {
		host = addr
	}

	if rt.originalHost == "" {
		rt.originalHost = host
	}

	conn := utls.UClient(rawConn, &utls.Config{ServerName: rt.originalHost, VerifyPeerCertificate: VerifyCert, InsecureSkipVerify: true}, rt.clientHelloId)
	if err = conn.Handshake(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	if rt.cachedTransports[addr] != nil {
		return conn, nil
	}

	// No http.Transport constructed yet, create one based on the results
	// of ALPN.
	switch conn.ConnectionState().NegotiatedProtocol {
	case http2.NextProtoTLS:
		// The remote peer is speaking HTTP 2 + TLS.
		rt.cachedTransports[addr] = &http2.Transport{DialTLS: rt.dialTLSHTTP2}
	default:
		// Assume the remote peer is speaking HTTP 1.x + TLS.
		rt.cachedTransports[addr] = &http.Transport{DialTLSContext: rt.dialTLS}
	}

	// Stash the connection just established for use servicing the
	// actual request (should be near-immediate).
	rt.cachedConnections[addr] = conn

	return nil, errProtocolNegotiated
}

func (rt *roundTripper) dialTLSHTTP2(network, addr string, _ *tls.Config) (net.Conn, error) {
	return rt.dialTLS(context.Background(), network, addr)
}

func (rt *roundTripper) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(req.URL.Host, "443") // we can assume port is 443 at this point
}

func newRoundTripper(clientHello utls.ClientHelloID, dialer ...proxy.ContextDialer) http.RoundTripper {
	if len(dialer) > 0 {
		return &roundTripper{
			dialer: dialer[0],

			clientHelloId: clientHello,

			cachedTransports:  make(map[string]http.RoundTripper),
			cachedConnections: make(map[string]net.Conn),
		}
	} else {
		return &roundTripper{
			dialer: proxy.Direct,

			clientHelloId: clientHello,

			cachedTransports:  make(map[string]http.RoundTripper),
			cachedConnections: make(map[string]net.Conn),
		}
	}
}

func VerifyCert(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {

	var fingerprints []string

	for _, rawCert := range rawCerts {
		fingerprints = append(fingerprints, fmt.Sprintf("%x", sha256.Sum256(rawCert)))
	}

	for _, fingerprint := range fingerprints {

		//log.Println(fingerprint)
		// Check to see if the fingerprints match for the site
		for f := range SitePrints {
			if fingerprint == SitePrints[f][1] {
				return nil
			}
		}
	}
	//return nil
	return errors.New("client error with the request - verify sig")
}
