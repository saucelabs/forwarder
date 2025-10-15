// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"fmt"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/saucelabs/forwarder/log"
)

type KerberosConfig struct {
	Enabled        bool
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

func (a *KerberosAdapter) connectToKDC() error {
	a.log.Debug("Logging to KDC server")
	err := a.krb5client.Login()
	if err != nil {
		return fmt.Errorf("kerberos KDC login: %w", err)
	}

	a.log.Debug("KDC login successful")

	return nil
}
