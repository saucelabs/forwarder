// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/saucelabs/forwarder/utils/certutil"
)

type TLSClientConfig struct {
	// HandshakeTimeout specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	HandshakeTimeout time.Duration

	// InsecureSkipVerify controls whether a client verifies the server's
	// certificate chain and host name. If InsecureSkipVerify is true, crypto/tls
	// accepts any certificate presented by the server and any host name in that
	// certificate. In this mode, TLS is susceptible to machine-in-the-middle
	// attacks unless custom verification is used. This should be used only for
	// testing or in combination with VerifyConnection or VerifyPeerCertificate.
	InsecureSkipVerify bool

	// CACertFiles is a list of paths to CA certificate files.
	// If this is set, the system root CA pool will be supplemented with certificates from these files.
	CACertFiles []string

	// KeyLogFile optionally specifies a destination for TLS master secrets
	// in NSS key log format that can be used to allow external programs
	// such as Wireshark to decrypt TLS connections.
	KeyLogFile string
}

func DefaultTLSClientConfig() *TLSClientConfig {
	return &TLSClientConfig{
		HandshakeTimeout: 10 * time.Second,
		KeyLogFile:       os.Getenv("SSLKEYLOGFILE"),
	}
}

func (c *TLSClientConfig) ConfigureTLSConfig(tlsCfg *tls.Config) error {
	tlsCfg.InsecureSkipVerify = c.InsecureSkipVerify

	if err := c.loadRootCAs(tlsCfg); err != nil {
		return fmt.Errorf("load CAs: %w", err)
	}

	if c.KeyLogFile != "" {
		f, err := os.OpenFile(c.KeyLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return fmt.Errorf("open key log file: %w", err)
		}
		tlsCfg.KeyLogWriter = f
	}

	return nil
}

func (c *TLSClientConfig) loadRootCAs(tlsCfg *tls.Config) error {
	if len(c.CACertFiles) == 0 {
		return nil
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return err
	}

	for _, name := range c.CACertFiles {
		b, err := ReadFileOrBase64(name)
		if err != nil {
			return err
		}
		if !rootCAs.AppendCertsFromPEM(b) {
			return fmt.Errorf("append certificate %q", name)
		}
	}

	tlsCfg.RootCAs = rootCAs

	return nil
}

type TLSServerConfig struct {
	// HandshakeTimeout specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	HandshakeTimeout time.Duration

	// CertFile is the path to the TLS certificate.
	CertFile string

	// KeyFile is the path to the TLS private key of the certificate.
	KeyFile string
}

func (c *TLSServerConfig) ConfigureTLSConfig(tlsCfg *tls.Config) error {
	if err := c.loadCertificate(tlsCfg); err != nil {
		return fmt.Errorf("load certificate: %w", err)
	}

	return nil
}

func (c *TLSServerConfig) loadCertificate(tlsCfg *tls.Config) error {
	var (
		cert tls.Certificate
		err  error
	)

	if c.CertFile == "" && c.KeyFile == "" {
		ssc := certutil.ECDSASelfSignedCert()

		if n, err := os.Hostname(); err == nil {
			ssc.Hosts = append(ssc.Hosts, n)
		}
		ssc.Hosts = append(ssc.Hosts, "localhost")

		cert, err = ssc.Gen()
	} else {
		cert, err = loadX509KeyPair(c.CertFile, c.KeyFile)
	}

	if err == nil {
		tlsCfg.Certificates = append(tlsCfg.Certificates, cert)
	}
	return err
}

func loadX509KeyPair(certFile, keyFile string) (tls.Certificate, error) {
	certPEMBlock, err := ReadFileOrBase64(certFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEMBlock, err := ReadFileOrBase64(keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
}
