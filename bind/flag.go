// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package bind

import (
	"net/url"
	"os"
	"strings"

	"github.com/mmatczuk/anyflag"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/fileurl"
	"github.com/saucelabs/forwarder/httplog"
	"github.com/saucelabs/forwarder/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func DNSConfig(fs *pflag.FlagSet, cfg *forwarder.DNSConfig) {
	fs.VarP(anyflag.NewSliceValue[*url.URL](nil, &cfg.Servers, forwarder.ParseDNSAddress),
		"dns-server", "n", "DNS server IP or URL ex. 1.1.1.1 or udp://1.1.1.1:53 (can be specified multiple times)")
	fs.DurationVar(&cfg.Timeout,
		"dns-timeout", cfg.Timeout, "timeout for DNS queries if DNS server is specified")
}

func PAC(fs *pflag.FlagSet, pac **url.URL) {
	fs.VarP(anyflag.NewValue[*url.URL](*pac, pac, fileurl.ParseFilePathOrURL),
		"pac", "p", "local file `path or URL` to PAC content, use \"-\" to read from stdin")
}

func HTTPProxyConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPProxyConfig, lcfg *log.Config) {
	HTTPServerConfig(fs, &cfg.HTTPServerConfig, "", forwarder.HTTPScheme, forwarder.HTTPSScheme)
	LogConfig(fs, lcfg)
	fs.VarP(anyflag.NewValue[*url.URL](cfg.UpstreamProxy, &cfg.UpstreamProxy, forwarder.ParseProxyURL),
		"upstream-proxy", "u", "upstream proxy URL")

	proxyLocalhostValues := []forwarder.ProxyLocalhostMode{
		forwarder.DenyProxyLocalhost,
		forwarder.AllowProxyLocalhost,
		forwarder.DirectProxyLocalhost,
	}
	fs.VarP(anyflag.NewValue[forwarder.ProxyLocalhostMode](cfg.ProxyLocalhost, &cfg.ProxyLocalhost, anyflag.EnumParser[forwarder.ProxyLocalhostMode](proxyLocalhostValues...)),
		"proxy-localhost", "t", "accept or deny requests to localhost, one of deny, allow, direct; in direct mode localhost requests are not sent to upstream proxy if present")

	fs.StringSliceVar(&cfg.RemoveHeaders, "remove-headers", cfg.RemoveHeaders, "removes request headers if prefixes match (can be specified multiple times)")
}

func HTTPTransportConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPTransportConfig) {
	fs.DurationVar(&cfg.DialTimeout,
		"http-dial-timeout", cfg.DialTimeout, "dial timeout for HTTP connections")
	fs.DurationVar(&cfg.KeepAlive,
		"http-keep-alive", cfg.KeepAlive, "keep alive interval for HTTP connections")
	fs.DurationVar(&cfg.TLSHandshakeTimeout,
		"http-tls-handshake-timeout", cfg.TLSHandshakeTimeout, "TLS handshake timeout for HTTP connections")
	fs.IntVar(&cfg.MaxIdleConns,
		"http-max-idle-conns", cfg.MaxIdleConns, "maximum number of idle connections for HTTP connections")
	fs.IntVar(&cfg.MaxIdleConnsPerHost,
		"http-max-idle-conns-per-host", cfg.MaxIdleConnsPerHost, "maximum number of idle connections per host for HTTP connections")
	fs.IntVar(&cfg.MaxConnsPerHost,
		"http-max-conns-per-host", cfg.MaxConnsPerHost, "maximum number of connections per host for HTTP connections")
	fs.DurationVar(&cfg.IdleConnTimeout,
		"http-idle-conn-timeout", cfg.IdleConnTimeout, "idle connection timeout for HTTP connections")
	fs.DurationVar(&cfg.ResponseHeaderTimeout,
		"http-response-header-timeout", cfg.ResponseHeaderTimeout, "response header timeout for HTTP connections")
	fs.DurationVar(&cfg.ExpectContinueTimeout,
		"http-expect-continue-timeout", cfg.ExpectContinueTimeout, "expect continue timeout for HTTP connections")

	TLSConfig(fs, &cfg.TLSConfig)
}

func HTTPServerConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPServerConfig, prefix string, schemes ...forwarder.Scheme) {
	namePrefix := prefix
	if namePrefix != "" {
		namePrefix += "-"
	}

	usagePrefix := prefix
	if usagePrefix != "" {
		usagePrefix += " "
	}

	if schemes == nil {
		schemes = []forwarder.Scheme{
			forwarder.HTTPScheme,
			forwarder.HTTPSScheme,
			forwarder.HTTP2Scheme,
		}
	}

	supportedSchemesStr := func() string {
		var sb strings.Builder
		for _, s := range schemes {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(string(s))
		}
		return sb.String()
	}

	fs.VarP(anyflag.NewValue[forwarder.Scheme](cfg.Protocol, &cfg.Protocol,
		anyflag.EnumParser[forwarder.Scheme](schemes...)),
		namePrefix+"protocol", "", usagePrefix+"HTTP server protocol, one of "+supportedSchemesStr())
	fs.StringVarP(&cfg.Addr,
		namePrefix+"address", "", cfg.Addr, usagePrefix+"HTTP server listen address in the form of `host:port`")
	fs.StringVar(&cfg.CertFile,
		namePrefix+"cert-file", cfg.CertFile, usagePrefix+"HTTP server TLS certificate file")
	fs.StringVar(&cfg.KeyFile,
		namePrefix+"key-file", cfg.KeyFile, usagePrefix+"HTTP server TLS key file")
	fs.DurationVar(&cfg.ReadTimeout,
		namePrefix+"read-timeout", cfg.ReadTimeout, usagePrefix+"HTTP server read timeout")
	fs.DurationVar(&cfg.ReadHeaderTimeout,
		namePrefix+"read-header-timeout", cfg.ReadHeaderTimeout, usagePrefix+"HTTP server read header timeout")
	fs.DurationVar(&cfg.WriteTimeout,
		namePrefix+"write-timeout", cfg.WriteTimeout, usagePrefix+"HTTP server write timeout")
	fs.VarP(anyflag.NewValue[*url.Userinfo](cfg.BasicAuth, &cfg.BasicAuth, forwarder.ParseUserInfo),
		namePrefix+"basic-auth", "", usagePrefix+"HTTP server basic-auth in the form of `username:password`")
	fs.Var(anyflag.NewValue[httplog.LoggerMode](cfg.LogHTTPMode, &cfg.LogHTTPMode, httplog.ParseMode),
		namePrefix+"log-http-requests", usagePrefix+"log http request, one of url, headers, body, error; error mode is default and logs requests with status code >= 500")
}

func TLSConfig(fs *pflag.FlagSet, cfg *forwarder.TLSConfig) {
	fs.BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-verify", cfg.InsecureSkipVerify, "skip TLS verification")
}

func LogConfig(fs *pflag.FlagSet, cfg *log.Config) {
	fs.VarP(anyflag.NewValue[*os.File](nil, &cfg.File,
		forwarder.OpenFileParser(os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600, 0o700)),
		"log-file", "", "log file path (default: stdout)")
	fs.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "enable verbose logging")
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

func MarkFlagFilename(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagFilename(name); err != nil {
			panic(err)
		}
	}
}
