// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/krberror"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/saucelabs/forwarder/log"
)

type KerberosConfig struct {
	Enabled        bool
	RunDiagnostics bool
	CfgFilePath    string
	KeyTabFilePath string
	UserName       string
	UserRealm      string
	// no matching and wildcards like in MITMHosts
	KerberosEnabledHosts []string
}

func DefaultKerberosConfig() *KerberosConfig {
	return &KerberosConfig{
		// default zero values are fine
	}
}

type KerberosAdapter struct {
	configuration KerberosConfig
	krb5client    client.Client
	log           log.StructuredLogger
}

func NewKerberosAdapter(cnf KerberosConfig, log log.StructuredLogger) (*KerberosAdapter, error) {
	// technically this should not happen as adapter should not be initialized without
	// proper config present, but better safe than sorry
	if cnf.CfgFilePath == "" {
		return nil, fmt.Errorf("kerberos config file (krb5.conf) not specified")
	}

	if cnf.KeyTabFilePath == "" {
		return nil, fmt.Errorf("kerberos keytab file not specified")
	}

	krb5Config, err := config.Load(cnf.CfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("error loading kerberos config file %s: %w", cnf.CfgFilePath, err)
	}

	krb5Keytab, err := keytab.Load(cnf.KeyTabFilePath)
	if err != nil {
		return nil, fmt.Errorf("error loading kerberos keytab file %s: %w", cnf.KeyTabFilePath, err)
	}

	if cnf.UserName == "" {
		return nil, fmt.Errorf("kerberos username not specified")
	}
	if cnf.UserRealm == "" {
		return nil, fmt.Errorf("kerberos user realm not specified")
	}

	krb5Client := client.NewWithKeytab(cnf.UserName, cnf.UserRealm, krb5Keytab, krb5Config)

	return &KerberosAdapter{configuration: cnf, krb5client: *krb5Client, log: log}, nil
}

func (a *KerberosAdapter) ConnectToKDC() error {
	a.log.Debug("Logging to KDC server")
	loginErr := a.krb5client.Login()
	if loginErr != nil && !a.configuration.RunDiagnostics {
		return fmt.Errorf("kerberos KDC login: %w", loginErr)
	}

	if loginErr != nil && a.configuration.RunDiagnostics {
		a.log.Error("kerberos KDC login failed but running diagnostics anyway", "error", loginErr)
	}

	a.log.Info("Kerberos KDC login successful")

	// run diagnostics even if login failed
	if a.configuration.RunDiagnostics {
		a.log.Warn("Kerberos diagnostics mode - diagnostic info will be printed to stdout and forwarder process will exit.")
		buf := new(bytes.Buffer)
		err := a.krb5client.Diagnostics(buf)

		// We need to print directly to stdout as it contains a nested structured text.
		// Does not really matter as diagnostics mode should be used on local console only.
		fmt.Printf("%s", buf.String())

		if err != nil {
			return fmt.Errorf("kerberos configuration potential problems: %w", err)
		}

		return fmt.Errorf("no kerberos configuration problems found. Exiting process")

	}

	return nil
}

// GetSPNEGOHeaderValue accepts SPN service name and returns header value that should
// be put inside Authorization or Proxy-Authorization header.
func (a *KerberosAdapter) GetSPNEGOHeaderValue(spn string) (string, error) {
	a.log.Debug("Generating SPNEGO header value for SPN: %s", spn)

	cli := spnego.SPNEGOClient(&a.krb5client, spn)

	err := cli.AcquireCred()
	if err != nil {
		return "", fmt.Errorf("could not acquire SPNEGO client credential: %w", err)
	}

	secContext, err := cli.InitSecContext()
	if err != nil {
		return "", fmt.Errorf("could not initialize SPNEGO context for SPN %s: %w", spn, err)
	}
	nb, err := secContext.Marshal()
	if err != nil {
		return "", krberror.Errorf(err, krberror.EncodingError, "could not marshal SPNEGO")
	}
	return "Negotiate " + base64.StdEncoding.EncodeToString(nb), nil

}
