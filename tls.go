// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"crypto/tls"

	"github.com/saucelabs/forwarder/utils/certutil"
)

type TLSClientConfig struct {
	// InsecureSkipVerify controls whether a client verifies the server's
	// certificate chain and host name. If InsecureSkipVerify is true, crypto/tls
	// accepts any certificate presented by the server and any host name in that
	// certificate. In this mode, TLS is susceptible to machine-in-the-middle
	// attacks unless custom verification is used. This should be used only for
	// testing or in combination with VerifyConnection or VerifyPeerCertificate.
	InsecureSkipVerify bool
}

type TLSServerConfig struct {
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
		cert, err = certutil.RSASelfSignedCert().Gen()
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
