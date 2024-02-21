// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strings"

	"github.com/mmatczuk/anyflag"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/fileurl"
	"github.com/saucelabs/forwarder/header"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/ruleset"
	"github.com/saucelabs/forwarder/utils/osdns"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

func ConfigFile(fs *pflag.FlagSet, configFile *string) {
	fs.StringVarP(configFile,
		"config-file", "c", *configFile, "<path>"+
			"Configuration file to load options from. "+
			"The supported formats are: JSON, YAML, TOML, HCL, and Java properties. "+
			"The file format is determined by the file extension, if not specified the default format is YAML. "+
			"The following precedence order of configuration sources is used: command flags, environment variables, config file, default values. ")
}

func DNSConfig(fs *pflag.FlagSet, cfg *osdns.Config) {
	fs.VarP(anyflag.NewSliceValue[netip.AddrPort](nil, &cfg.Servers, forwarder.ParseDNSAddress),
		"dns-server", "n", "<ip>[:<port>]"+
			"DNS server(s) to use instead of system default. "+
			"There are two execution policies, when more then one server is specified. "+
			"Fallback: the first server in a list is used as primary, the rest are used as fallbacks. "+
			"Round robin: the servers are used in a round-robin fashion. "+
			"The port is optional, if not specified the default port is 53. ")

	fs.DurationVar(&cfg.Timeout,
		"dns-timeout", cfg.Timeout, "Timeout for dialing DNS servers. "+
			"Only used if DNS servers are specified. ")

	fs.BoolVar(&cfg.RoundRobin, "dns-round-robin", cfg.RoundRobin,
		"If more than one DNS server is specified with the --dns-server flag, "+
			"passing this flag will enable round-robin selection. ")
}

func PAC(fs *pflag.FlagSet, pac **url.URL) {
	fs.VarP(anyflag.NewValue[*url.URL](*pac, pac, fileurl.ParseFilePathOrURL),
		"pac", "p", "<path or URL>"+
			"Proxy Auto-Configuration file to use for upstream proxy selection. "+
			"It can be a local file or a URL, you can also use '-' to read from stdin. "+
			"The data URI scheme is supported, the format is data:base64,<encoded data>. ")
}

func ProxyHeaders(fs *pflag.FlagSet, headers *[]header.Header) {
	fs.Var(anyflag.NewSliceValueWithRedact[header.Header](*headers, headers, header.ParseHeader, RedactHeader),
		"proxy-header", "<header>"+
			"Add or remove HTTP headers on the CONNECT request to the upstream proxy. "+
			"See the documentation for the -H, --header flag for more details on the format. ")
}

func RequestHeaders(fs *pflag.FlagSet, headers *[]header.Header) {
	fs.VarP(anyflag.NewSliceValueWithRedact[header.Header](*headers, headers, header.ParseHeader, RedactHeader),
		"header", "H", "<header>"+
			"Add or remove HTTP request headers. "+
			"Use the format \"name: value\" to add a header, "+
			"\"name;\" to set the header to empty value, "+
			"\"-name\" to remove the header, "+
			"\"-name*\" to remove headers by prefix. "+
			"The header name will be normalized to canonical form. "+
			"The header value should not contain any newlines or carriage returns. "+
			"The flag can be specified multiple times. "+
			"Example: -H \"Host: example.com\" -H \"-User-Agent\" -H \"-X-*\". ")
}

func ResponseHeaders(fs *pflag.FlagSet, headers *[]header.Header) {
	fs.VarP(anyflag.NewSliceValueWithRedact[header.Header](*headers, headers, header.ParseHeader, RedactHeader),
		"response-header", "R", "<header>"+
			"Add or remove HTTP headers on the received response before sending it to the client. "+
			"See the documentation for the -H, --header flag for more details on the format. ")
}

func HTTPProxyConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPProxyConfig, lcfg *log.Config) {
	HTTPServerConfig(fs, &cfg.HTTPServerConfig, "", forwarder.HTTPScheme, forwarder.HTTPSScheme)
	LogConfig(fs, lcfg)

	fs.VarP(anyflag.NewValueWithRedact[*url.URL](cfg.UpstreamProxy, &cfg.UpstreamProxy, forwarder.ParseProxyURL, RedactURL),
		"proxy", "x", "[protocol://]host[:port]"+
			"Upstream proxy to use. "+
			"The supported protocols are: http, https, socks5. "+
			"No protocol specified will be treated as HTTP proxy. "+
			"If the port number is not specified, it is assumed to be 1080. "+
			"The basic authentication username and password can be specified in the host string e.g. user:pass@host:port. "+
			"Alternatively, you can use the -c, --credentials flag to specify the credentials. "+
			"If both are specified, the proxy flag takes precedence. ")

	proxyLocalhostValues := []forwarder.ProxyLocalhostMode{
		forwarder.DenyProxyLocalhost,
		forwarder.AllowProxyLocalhost,
		forwarder.DirectProxyLocalhost,
	}
	fs.VarP(anyflag.NewValue[forwarder.ProxyLocalhostMode](cfg.ProxyLocalhost, &cfg.ProxyLocalhost, anyflag.EnumParser[forwarder.ProxyLocalhostMode](proxyLocalhostValues...)),
		"proxy-localhost", "", "<allow|deny|direct>"+
			"Setting this to allow enables sending requests to localhost through the upstream proxy. "+
			"Setting this to direct sends requests to localhost directly without using the upstream proxy. "+
			"By default, requests to localhost are denied. ")

	fs.StringVar(&cfg.Name, "name", cfg.Name, "<string>"+
		"Name of this proxy instance. This value is used in the Via header in requests. "+
		"The name value in Via header is extended with a random string to avoid collisions when several proxies are chained. ")

	fs.StringVar(&cfg.RequestIDHeader, "log-http-request-id-header", cfg.RequestIDHeader,
		"<name>"+
			"If the header is present in the request, "+
			"the proxy will associate the value with the request in the logs. ")

	fs.Var(&cfg.ReadLimit, "read-limit", "<bandwidth>"+
		"Global read rate limit in bytes per second i.e. how many bytes per second you can receive from a proxy. "+
		"Accepts binary format (e.g. 1.5Ki, 1Mi, 3.6Gi). ")

	fs.Var(&cfg.WriteLimit, "write-limit", "<bandwidth>"+
		"Global write rate limit in bytes per second i.e. how many bytes per second you can send to proxy. "+
		"Accepts binary format (e.g. 1.5Ki, 1Mi, 3.6Gi). ")
}

func DenyDomains(fs *pflag.FlagSet, cfg *[]ruleset.RegexpListItem) {
	fs.Var(anyflag.NewSliceValue[ruleset.RegexpListItem](*cfg, cfg, ruleset.ParseRegexpListItem),
		"deny-domains", "[-]<regexp>,..."+
			"Deny requests to the specified domains. "+
			"Prefix domains with '-' to exclude requests to certain domains from being denied.")
}

func DirectDomains(fs *pflag.FlagSet, cfg *[]ruleset.RegexpListItem) {
	fs.Var(anyflag.NewSliceValue[ruleset.RegexpListItem](*cfg, cfg, ruleset.ParseRegexpListItem),
		"direct-domains", "[-]<regexp>,..."+
			"Connect directly to the specified domains without using the upstream proxy. "+
			"Prefix domains with '-' to exclude requests to certain domains from being directed. "+
			"This flag takes precedence over the PAC script.")
}

func MITMConfig(fs *pflag.FlagSet, mitm *bool, cfg *forwarder.MITMConfig) {
	fs.BoolVar(mitm, "mitm", *mitm, ""+
		"Enable Man-in-the-Middle (MITM) mode. "+
		"It only works with HTTPS requests, HTTP/2 is not supported. "+
		"MITM is enabled by default when the --mitm-cacert-file flag is set. "+
		"If the CA certificate is not provided MITM uses a generated CA certificate. "+
		"The CA certificate used can be retrieved from the API server .")

	fs.Var(anyflag.NewValueWithRedact[string](cfg.CACertFile, &cfg.CACertFile, func(val string) (string, error) { return val, nil }, RedactBase64),
		"mitm-cacert-file", "<path or base64>"+
			"CA certificate file to use for generating MITM certificates. "+
			"If the file is not specified, a generated CA certificate will be used. "+
			"See the documentation for the --mitm flag for more details. ")

	fs.Var(anyflag.NewValueWithRedact[string](cfg.CAKeyFile, &cfg.CAKeyFile, func(val string) (string, error) { return val, nil }, RedactBase64),
		"mitm-cakey-file", "<path or base64>"+
			"CA key file to use for generating MITM certificates. ")

	fs.StringVar(&cfg.Organization, "mitm-org", cfg.Organization, "<name>"+
		"Organization name to use in the generated MITM certificates. ")

	fs.DurationVar(&cfg.Validity, "mitm-validity", cfg.Validity, ""+
		"Validity period of the generated MITM certificates. ")
}

func MITMDomains(fs *pflag.FlagSet, cfg *[]ruleset.RegexpListItem) {
	fs.Var(anyflag.NewSliceValue[ruleset.RegexpListItem](*cfg, cfg, ruleset.ParseRegexpListItem),
		"mitm-domains", "[-]<regexp>,..."+
			"Limit MITM to the specified domains. "+
			"Prefix domains with '-' to exclude requests to certain domains from being MITMed.")
}

func Credentials(fs *pflag.FlagSet, credentials *[]*forwarder.HostPortUser) {
	fs.VarP(anyflag.NewSliceValueWithRedact[*forwarder.HostPortUser](*credentials, credentials, forwarder.ParseHostPortUser, forwarder.RedactHostPortUser),
		"credentials", "s", "<username[:password]@host:port,...>"+
			"Site or upstream proxy basic authentication credentials. "+
			"The host and port can be set to \"*\" to match all hosts and ports respectively. "+
			"The flag can be specified multiple times to add multiple credentials. ")
}

func HTTPTransportConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPTransportConfig) {
	DialConfig(fs, &cfg.DialConfig, "http")

	TLSClientConfig(fs, &cfg.TLSClientConfig)

	fs.DurationVar(&cfg.IdleConnTimeout,
		"http-idle-conn-timeout", cfg.IdleConnTimeout,
		"The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. "+
			"Zero means no limit. ")

	fs.DurationVar(&cfg.ResponseHeaderTimeout,
		"http-response-header-timeout", cfg.ResponseHeaderTimeout,
		"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any)."+
			"This time does not include the time to read the response body. "+
			"Zero means no limit. ")
}

func DialConfig(fs *pflag.FlagSet, cfg *forwarder.DialConfig, prefix string) {
	namePrefix := prefix
	if namePrefix != "" {
		namePrefix += "-"
	}

	fs.DurationVar(&cfg.DialTimeout,
		namePrefix+"dial-timeout", cfg.DialTimeout,
		"The maximum amount of time a dial will wait for a connect to complete. "+
			"With or without a timeout, the operating system may impose its own earlier timeout. For instance, TCP timeouts are often around 3 minutes. ")
}

func TLSClientConfig(fs *pflag.FlagSet, cfg *forwarder.TLSClientConfig) {
	fs.DurationVar(&cfg.HandshakeTimeout,
		"http-tls-handshake-timeout", cfg.HandshakeTimeout,
		"The maximum amount of time waiting to wait for a TLS handshake. Zero means no limit.")

	fs.BoolVar(&cfg.InsecureSkipVerify, "insecure", cfg.InsecureSkipVerify,
		"Don't verify the server's certificate chain and host name. "+
			"Enable to work with self-signed certificates. ")

	fs.Var(anyflag.NewSliceValueWithRedact[string](cfg.CACertFiles, &cfg.CACertFiles, func(val string) (string, error) { return val, nil }, RedactBase64),
		"cacert-file", "<path or base64>"+
			"Add your own CA certificates to verify against. "+
			"The system root certificates will be used in addition to any certificates in this list. "+
			"Can be a path to a file or \"data:\" followed by base64 encoded certificate. "+
			"Use this flag multiple times to specify multiple CA certificate files. ")
}

func HTTPServerConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPServerConfig, prefix string, schemes ...forwarder.Scheme) {
	namePrefix := prefix
	if namePrefix != "" {
		namePrefix += "-"
	}

	fs.StringVarP(&cfg.Addr,
		namePrefix+"address", "", cfg.Addr, "<host:port>"+
			"The server address to listen on. "+
			"If the host is empty, the server will listen on all available interfaces. ")

	if schemes == nil {
		schemes = []forwarder.Scheme{
			forwarder.HTTPScheme,
			forwarder.HTTPSScheme,
			forwarder.HTTP2Scheme,
		}
	}

	if len(schemes) > 1 {
		supportedSchemesStr := func(delim string) string {
			var sb strings.Builder
			for _, s := range schemes {
				if sb.Len() > 0 {
					sb.WriteString(delim)
				}
				sb.WriteString(string(s))
			}
			return sb.String()
		}

		fs.VarP(anyflag.NewValue[forwarder.Scheme](cfg.Protocol, &cfg.Protocol,
			anyflag.EnumParser[forwarder.Scheme](schemes...)),
			namePrefix+"protocol", "", "<"+supportedSchemesStr("|")+">"+
				"The server protocol. "+
				"For https and h2 protocols, if TLS certificate is not specified, "+
				"the server will use a self-signed certificate. ")

		TLSServerConfig(fs, &cfg.TLSServerConfig, namePrefix)
	}

	fs.DurationVar(&cfg.IdleTimeout, namePrefix+"idle-timeout", cfg.IdleTimeout,
		"The maximum amount of time to wait for the next request before closing connection.")

	fs.DurationVar(&cfg.ReadHeaderTimeout,
		namePrefix+"read-header-timeout", cfg.ReadHeaderTimeout,
		"The amount of time allowed to read request headers.")

	fs.VarP(anyflag.NewValueWithRedact[*url.Userinfo](cfg.BasicAuth, &cfg.BasicAuth, forwarder.ParseUserinfo, RedactUserinfo),
		namePrefix+"basic-auth", "", "<username[:password]>"+
			"Basic authentication credentials to protect the server. ")
}

func HTTPLogConfig(fs *pflag.FlagSet, cfg []NamedParam[httplog.Mode]) {
	for _, p := range cfg {
		if p.Param == nil {
			panic("httplog mode is nil for " + p.Name)
		}
	}

	names := httplogExtractNames(cfg)

	parse := func(val string) (NamedParam[httplog.Mode], error) {
		name, mode, err := httplog.SplitNameMode(val)
		if err != nil {
			return NamedParam[httplog.Mode]{}, err
		}
		if name != "" && !slices.Contains(names, name) {
			return NamedParam[httplog.Mode]{}, fmt.Errorf("unknown name: %s", name)
		}

		return NamedParam[httplog.Mode]{Name: name, Param: &mode}, nil
	}

	var flagValue []NamedParam[httplog.Mode]
	f := httplogFlag{
		SliceValue: anyflag.NewSliceValue[NamedParam[httplog.Mode]](flagValue, &flagValue, parse),
		update: func() {
			httplogUpdate(cfg, flagValue)
		},
	}

	valueType := "<none|short-url|url|headers|body|errors>"
	if ss := names; len(ss) > 1 {
		valueType = "[" + strings.Join(ss, "|") + ":]" + valueType
	}

	fs.Var(f, "log-http", valueType+",... "+
		"HTTP request and response logging mode. "+
		"Setting this to none disables logging. "+
		"The short-url mode logs [scheme://]host[/path] instead of the full URL. "+
		"The error mode logs request line and headers if status code is greater than or equal to 500. ")
}

func TLSServerConfig(fs *pflag.FlagSet, cfg *forwarder.TLSServerConfig, namePrefix string) {
	fs.DurationVar(&cfg.HandshakeTimeout,
		namePrefix+"tls-handshake-timeout", cfg.HandshakeTimeout,
		"The maximum amount of time to wait for a TLS handshake before closing connection. Zero means no limit.")

	fs.Var(anyflag.NewValueWithRedact[string](cfg.CertFile, &cfg.CertFile, func(val string) (string, error) { return val, nil }, RedactBase64),
		namePrefix+"tls-cert-file", "<path or base64>"+
			"TLS certificate to use if the server protocol is https or h2. "+
			"Can be a path to a file or \"data:\" followed by base64 encoded certificate. ")

	fs.Var(anyflag.NewValueWithRedact[string](cfg.KeyFile, &cfg.KeyFile, func(val string) (string, error) { return val, nil }, RedactBase64),
		namePrefix+"tls-key-file", "<path or base64>"+
			"TLS private key to use if the server protocol is https or h2. "+
			"Can be a path to a file or \"data:\" followed by base64 encoded key. ")
}

func LogConfig(fs *pflag.FlagSet, cfg *log.Config) {
	fs.VarP(newOSFileFlag(anyflag.NewValue[*os.File](nil, &cfg.File,
		forwarder.OpenFileParser(os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600, 0o700)), &cfg.File),
		"log-file", "", "<path>"+
			"Path to the log file, if empty, logs to stdout. ")

	logLevel := []log.Level{
		log.ErrorLevel,
		log.InfoLevel,
		log.DebugLevel,
	}
	fs.Var(anyflag.NewValue[log.Level](cfg.Level, &cfg.Level, anyflag.EnumParser[log.Level](logLevel...)),
		"log-level", "<error|info|debug>"+
			"Log level. ")
}

func MarkFlagHidden(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.Flags().MarkHidden(name); err != nil {
			panic(err)
		}
	}
}

func MarkFlagRequired(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(err)
		}
	}
}

func AutoMarkFlagFilename(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if strings.HasPrefix(f.Usage, "<path") ||
			strings.HasSuffix(f.Name, "-file") ||
			strings.HasSuffix(f.Name, "-dir") {
			MarkFlagFilename(cmd, f.Name)
		}
	})
}

func MarkFlagFilename(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagFilename(name); err != nil {
			panic(err)
		}
	}
}
