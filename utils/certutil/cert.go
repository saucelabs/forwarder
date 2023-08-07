// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package certutil

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"
)

// SelfSignedCert specifies a self-signed certificate to be generated.
type SelfSignedCert struct {
	Hosts        []string
	Organization []string
	ValidFrom    time.Time
	ValidFor     time.Duration
	IsCA         bool
	RsaBits      int
	EcdsaCurve   string
	Ed25519Key   bool
}

func RSASelfSignedCert() *SelfSignedCert {
	return &SelfSignedCert{
		Hosts:        []string{"localhost"},
		Organization: []string{"Sauce Labs Inc."},
		ValidFrom:    time.Now(),
		ValidFor:     365 * 24 * time.Hour,
		RsaBits:      2048,
	}
}

func ECDSASelfSignedCert() *SelfSignedCert {
	return &SelfSignedCert{
		Hosts:        []string{"localhost"},
		Organization: []string{"Sauce Labs Inc."},
		ValidFrom:    time.Now(),
		ValidFor:     365 * 24 * time.Hour,
		EcdsaCurve:   "P256",
		Ed25519Key:   true,
	}
}

// Gen generates a self-signed certificate, the implementation is based on https://golang.org/src/crypto/tls/generate_cert.go.
func (c *SelfSignedCert) Gen() (tls.Certificate, error) {
	var cert tls.Certificate

	priv, err := c.generateKey()
	if err != nil {
		return cert, fmt.Errorf("generate private key %w", err)
	}

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return cert, fmt.Errorf("generate serial number %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: c.Organization,
		},
		NotBefore: c.ValidFrom,
		NotAfter:  c.ValidFrom.Add(c.ValidFor),

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range c.Hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if c.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return cert, fmt.Errorf("create certificate %w", err)
	}
	cert.Certificate = [][]byte{derBytes}
	cert.PrivateKey = priv

	return cert, nil
}

func (c *SelfSignedCert) generateKey() (priv any, err error) {
	switch c.EcdsaCurve {
	case "":
		if c.Ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, c.RsaBits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		err = fmt.Errorf("unrecognized elliptic curve: %q", c.EcdsaCurve)
	}
	return priv, err
}

func publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey) //nolint:forcetypeassert // this is the only way to get the public key
	default:
		return nil
	}
}
