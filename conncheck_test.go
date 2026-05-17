package conncheck_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"testing"
	"time"

	"github.com/murfffi/conncheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo(t *testing.T) {

	t.Run("plain", func(t *testing.T) {
		testSocketConn(t, nil)
	})
	t.Run("tls", func(t *testing.T) {
		tlsCert := randomTLSCertificate(t)
		testSocketConn(t, tlsCert)
	})
}

func testSocketConn(t *testing.T, tlsCert *tls.Certificate) {

	accepted := make(chan net.Conn)

	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, ln.Close())
	}()

	if tlsCert != nil {
		ln = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{*tlsCert}})
	}
	go func() {
		for {
			cn, err := ln.Accept()
			if err != nil {
				// This is usually caused by Listener being
				// closed, not really an error.
				t.Log("accept goroutine completed")
				return
			}
			accepted <- cn
		}
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if tlsCert != nil {
		conn = tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
	}

	require.Equal(t, conncheck.StatusOpen, conncheck.Do(conn))

	serverConn := <-accepted
	go func() {
		_, err := serverConn.Write([]byte("hello"))
		assert.NoError(t, err)
	}()

	out := [5]byte{}
	n, err := conn.Read(out[0:1])
	require.NoError(t, err)
	require.Equal(t, 1, n)

	require.Equal(t, conncheck.StatusOpen, conncheck.Do(conn))

	n, err = conn.Read(out[1:])
	require.NoError(t, err)
	require.Equal(t, 4, n)

	require.Equal(t, conncheck.StatusOpen, conncheck.Do(conn))
	require.Equal(t, "hello", string(out[:]))

	if serverTLSConn, ok := serverConn.(*tls.Conn); ok {
		// in case of TLS we need to additionally ensure that the connection closes immediately
		tcpConn := serverTLSConn.NetConn().(*net.TCPConn)
		require.NoError(t, tcpConn.SetLinger(0))
		require.NoError(t, tcpConn.Close())
	} else {
		require.NoError(t, serverConn.Close())
	}
	require.Equal(t, conncheck.StatusNotOpen, conncheck.Do(conn))
}

func randomTLSCertificate(t *testing.T) *tls.Certificate {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now().Add(-time.Minute),
		NotAfter:  time.Now().Add(time.Hour),

		BasicConstraintsValid: true,

		DNSNames: []string{"localhost"},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	require.NoError(t, err)

	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privateKey,
	}

	return &cert
}
