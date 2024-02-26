// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"

	"github.com/saucelabs/forwarder/internal/martian/mitm"
	"github.com/saucelabs/forwarder/utils/certutil"
)

type MITMConfig struct {
	CACertFile string
	CAKeyFile  string

	Organization string
	Validity     time.Duration
}

func DefaultMITMConfig() *MITMConfig {
	return &MITMConfig{
		Organization: "Forwarder Proxy MITM",
		Validity:     24 * time.Hour, //nolint:gomnd // 24 hours is a reasonable default
	}
}

func (c *MITMConfig) loadCACertificate() (cert tls.Certificate, err error) {
	if c.CACertFile == "" && c.CAKeyFile == "" {
		tmpl := certutil.ECDSASelfSignedCert()
		tmpl.Organization = []string{c.Organization}
		tmpl.Hosts = nil
		tmpl.IsCA = true
		return tmpl.Gen()
	}

	return loadX509KeyPair(c.CACertFile, c.CAKeyFile)
}

func newMartianMITMConfig(c *MITMConfig) (*mitm.Config, error) {
	cert, err := c.loadCACertificate()
	if err != nil {
		return nil, err
	}
	ca, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	if !ca.IsCA {
		return nil, errors.New("certificate is not a CA")
	}

	cfg, err := mitm.NewConfig(ca, cert.PrivateKey)
	if err != nil {
		return nil, err
	}
	cfg.SetOrganization(c.Organization)
	cfg.SetValidity(c.Validity)

	return cfg, nil
}
