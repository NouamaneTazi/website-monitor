package main

import (
	"crypto/tls"
	"log"
	"net/http/httptrace"
	"time"
)

// HTTPTrace provides a datastructure for storing an http trace.
type HTTPTrace struct {
	DNSStart             time.Time
	ConnectStart         time.Time
	ConnectDone          time.Time
	GotConn              time.Time
	GotFirstResponseByte time.Time
	TLSHandshakeStart    time.Time
	TLSHandshakeDone     time.Time
	GotResponseBody      time.Time
}

// DNSLookup calculates duration of dns lookup
func (ht *HTTPTrace) DNSLookup() time.Duration {
	return ht.ConnectStart.Sub(ht.DNSStart)
}

// TCPConnection calculates duration of
func (ht *HTTPTrace) TCPConnection() time.Duration {
	return ht.ConnectDone.Sub(ht.ConnectStart)
}

// TLSHandshake calculates duration of
func (ht *HTTPTrace) TLSHandshake() time.Duration {
	return ht.TLSHandshakeDone.Sub(ht.TLSHandshakeStart)
}

// ServerProcessing calculates duration of
func (ht *HTTPTrace) ServerProcessing() time.Duration {
	return ht.GotFirstResponseByte.Sub(ht.GotConn)
}

// ContentTransfer calculates duration of
func (ht *HTTPTrace) ContentTransfer() time.Duration {
	return ht.GotResponseBody.Sub(ht.GotFirstResponseByte)
}

// NameLookup calculates duration of
func (ht *HTTPTrace) NameLookup() time.Duration {
	return ht.ConnectStart.Sub(ht.DNSStart)
}

// Connect calculates duration of
func (ht *HTTPTrace) Connect() time.Duration {
	return ht.ConnectDone.Sub(ht.DNSStart)
}

// PreTransfer calculates duration of
func (ht *HTTPTrace) PreTransfer() time.Duration {
	return ht.GotConn.Sub(ht.DNSStart)
}

// StartTransfer calculates duration of
func (ht *HTTPTrace) StartTransfer() time.Duration {
	return ht.GotFirstResponseByte.Sub(ht.DNSStart)
}

// Total calculates duration of
func (ht *HTTPTrace) Total() time.Duration {
	return ht.GotResponseBody.Sub(ht.DNSStart)
}

// trace returns a httptrace.ClientTrace object to be used with an http
// request via httptrace.WithClientTrace() that fills in the HttpTrace.
func (ht *HTTPTrace) trace() *httptrace.ClientTrace {

	trace := &httptrace.ClientTrace{
		// DNSStart is called when a DNS lookup begins.
		DNSStart: func(_ httptrace.DNSStartInfo) { ht.DNSStart = time.Now() },
		// DNSDone is called when a DNS lookup ends.
		DNSDone: func(_ httptrace.DNSDoneInfo) { ht.ConnectStart = time.Now() },

		// ConnectStart is called when a new connection's Dial begins.
		// If net.Dialer.DualStack (IPv6 "Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectStart: func(_, _ string) {
			if ht.DNSStart.IsZero() {
				// we skipped DNS
				ht.DNSStart = time.Now()
			}
			if ht.ConnectStart.IsZero() {
				// connecting to IP
				ht.ConnectStart = time.Now()
			}
		},

		// ConnectDone is called when a new connection's Dial
		// completes. The provided err indicates whether the
		// connection completedly successfully.
		// If net.Dialer.DualStack ("Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			ht.ConnectDone = time.Now()
			// println("Connected to ", addr)
		},

		// GetConn is called before a connection is created or
		// retrieved from an idle pool. The hostPort is the
		// "host:port" of the target or proxy. GetConn is called even
		// if there's already an idle cached connection available.
		GotConn: func(_ httptrace.GotConnInfo) { ht.GotConn = time.Now() },

		// GotFirstResponseByte is called when the first byte of the response
		// headers is available.
		GotFirstResponseByte: func() { ht.GotFirstResponseByte = time.Now() },

		// TLSHandshakeStart is called when the TLS handshake is started. When
		// connecting to an HTTPS site via an HTTP proxy, the handshake happens
		// after the CONNECT request is processed by the proxy.
		TLSHandshakeStart: func() { ht.TLSHandshakeStart = time.Now() },

		// TLSHandshakeDone is called after the TLS handshake with either the
		// successful handshake's connection state, or a non-nil error on handshake
		// failure.
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) { ht.TLSHandshakeDone = time.Now() },
	}

	return trace
}
