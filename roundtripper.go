package tlsclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/KakashiHatake324/tlsclient/v2/net/http2"
	"golang.org/x/net/proxy"

	utls "github.com/KakashiHatake324/tlsclient/v2/utls"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type CustomizedSettings struct {
	WriteData            bool
	MaxHeaderListSize    int
	ServerPushSet        bool
	ServerPushEnable     bool
	Priority             bool
	PriorityWeight       int // from 0 to 256
	InitialWindowSize    int
	MaxConcurrentStreams int
	HeaderTableSize      int
	WindowSizeIncrement  int
}

type roundTripper struct {
	sync.Mutex

	clientHelloId utls.ClientHelloID

	cachedConnections map[string]net.Conn
	cachedTransports  map[string]http.RoundTripper

	dialer       proxy.ContextDialer
	originalHost string
	cs           CustomizedSettings
	domain       string
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

	conn := utls.UClient(rawConn, &utls.Config{
		//KeyLogWriter:          w,
		ServerName:            rt.originalHost,
		VerifyPeerCertificate: VerifyCert,
		InsecureSkipVerify:    true,
	},
		rt.clientHelloId,
	)
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

		// chrome latest general settings
		// rt.cachedTransports[addr] = &http2.Transport{
		// 	DialTLS:              rt.dialTLSHTTP2,
		// 	MaxHeaderListSize:    262144,
		// 	ServerPushSet:        false,
		// 	ServerPushEnable:     false,
		// 	Priority:             true,
		// 	PriorityWeight:       255, // from 0 to 256
		// 	InitialWindowSize:    6291456,
		// 	MaxConcurrentStreams: 1000,
		// 	HeaderTableSize:      65536,
		// 	WindowSizeIncrement:  15663105,
		// }

		rt.cachedTransports[addr] = &http2.Transport{
			DialTLS:              rt.dialTLSHTTP2,
			MaxHeaderListSize:    rt.cs.MaxHeaderListSize,
			ServerPushSet:        rt.cs.ServerPushSet,
			ServerPushEnable:     rt.cs.ServerPushEnable,
			Priority:             rt.cs.Priority,
			PriorityWeight:       rt.cs.PriorityWeight,
			InitialWindowSize:    rt.cs.InitialWindowSize,
			MaxConcurrentStreams: rt.cs.MaxConcurrentStreams,
			HeaderTableSize:      rt.cs.HeaderTableSize,
			WindowSizeIncrement:  rt.cs.WindowSizeIncrement,
			WriteData:            rt.cs.WriteData,
		}

	default:
		// Assume the remote peer is speaking HTTP 1.x + TLS.
		rt.cachedTransports[addr] = &http.Transport{
			DialTLSContext: rt.dialTLS,
		}
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

func newRoundTripper(clientHello utls.ClientHelloID, settings CustomizedSettings, domain string, dialer ...proxy.ContextDialer) http.RoundTripper {
	if len(dialer) > 0 {
		return &roundTripper{
			dialer: dialer[0],

			clientHelloId: clientHello,

			cachedTransports:  make(map[string]http.RoundTripper),
			cachedConnections: make(map[string]net.Conn),
			cs:                settings,
		}
	} else {
		return &roundTripper{
			dialer: proxy.Direct,

			clientHelloId: clientHello,

			cachedTransports:  make(map[string]http.RoundTripper),
			cachedConnections: make(map[string]net.Conn),
			cs:                settings,
		}
	}
}

func VerifyCert(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {

	var domains []string

	for _, verifiedChain := range verifiedChains {
		for _, chain := range verifiedChain {
			for _, chainURL := range chain.DNSNames {
				if chainURL != "" {
					domains = append(domains, chainURL)
				}
			}
		}
	}

	log.Println(domains)

	for _, rawCert := range rawCerts {
		for _, domain := range domains {
			if loadedCerts[domain] == string(rawCert) {
				return nil
			}
		}
	}
	return errors.New("stop sniffing please")
}
