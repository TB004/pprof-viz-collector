// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shutdown gracefully stops srv within a short timeout and fails the test if
// that doesn't succeed. Intended for t.Cleanup.
func shutdown(t *testing.T, srv *Server) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assert.NoError(t, srv.Shutdown(ctx))
}

func generateTestCert(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	require.NoError(t, err)

	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")

	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}), 0o600))

	return certFile, keyFile
}

func TestNew(t *testing.T) {
	t.Parallel()

	server, err := New(0, NewConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, server) })
	assert.NotEmpty(t, server.Addr())

	protos := new(http.Protocols)
	protos.SetUnencryptedHTTP2(true)
	client := &http.Client{Transport: &http.Transport{Protocols: protos}}

	body := strings.NewReader("{}")
	req, err := http.NewRequest("POST", "http://"+server.Addr()+"/proto.CollectorService/SnapshotAll", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer assert.NoError(t, resp.Body.Close())
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNew_PortZeroPicksFreePort(t *testing.T) {
	t.Parallel()

	a, err := New(0, NewConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, a) })

	b, err := New(0, NewConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, b) })

	assert.NotEmpty(t, a.Addr())
	assert.NotEmpty(t, b.Addr())
	assert.NotEqual(t, a.Addr(), b.Addr(), "two servers on port 0 should get distinct ports")
}

func TestNew_Shutdown(t *testing.T) {
	t.Parallel()

	server, err := New(0, NewConfig{})
	require.NoError(t, err)
	addr := server.Addr()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, server.Shutdown(ctx))

	// Idempotent: a second call must not error or hang.
	assert.NoError(t, server.Shutdown(ctx))

	// The listener must actually be closed.
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if conn != nil {
		_ = conn.Close()
	}
	assert.Error(t, err)
}

func TestNew_HTTP3(t *testing.T) {
	t.Parallel()

	certFile, keyFile := generateTestCert(t)
	server, err := New(0, NewConfig{
		CertFile: certFile,
		KeyFile:  keyFile,
		HTTP3:    true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, server) })
	require.NotEmpty(t, server.Addr())

	conn, err := tls.Dial("tcp", server.Addr(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // test dial, not a real client
	assert.NoError(t, err)
	if conn != nil {
		_ = conn.Close()
	}
}

func TestNew_HTTP3RequiresCertFiles(t *testing.T) {
	t.Parallel()
	server, err := New(0, NewConfig{HTTP3: true})
	assert.Nil(t, server)
	assert.EqualError(t, err, "HTTP3 requires CertFile and KeyFile to be configured")
}
