package bind

import (
	"net/url"
	"os"

	"github.com/mmatczuk/anyflag"
	"github.com/saucelabs/forwarder"
	"github.com/saucelabs/forwarder/fileurl"
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

func HTTPProxyConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPProxyConfig) {
	fs.VarP(anyflag.NewValue[*url.URL](cfg.UpstreamProxy, &cfg.UpstreamProxy, forwarder.ParseProxyURL),
		"upstream-proxy", "u", "upstream proxy URL")
	fs.VarP(anyflag.NewValue[*url.Userinfo](cfg.UpstreamProxyCredentials, &cfg.UpstreamProxyCredentials, forwarder.ParseUserInfo),
		"upstream-proxy-credentials", "", "upstream proxy credentials in the form of `username:password`, it overrides credentials embedded in upstream-proxy URL")
	fs.VarP(anyflag.NewValue[*url.URL](cfg.PAC, &cfg.PAC, url.ParseRequestURI),
		"pac", "p", "local file `path or URL` to PAC content")
	fs.StringSliceVarP(&cfg.PACProxiesCredentials, "pac-proxies-credentials", "d", cfg.PACProxiesCredentials,
		"PAC proxies credentials in URL format ex. http://user:pass@host:port (can be specified multiple times)")
	fs.StringSliceVar(&cfg.SiteCredentials, "site-credentials", cfg.SiteCredentials,
		"target site credentials")
	fs.BoolVarP(&cfg.ProxyLocalhost, "proxy-localhost", "t", cfg.ProxyLocalhost,
		"if set, will proxy localhost requests to an upstream proxy")

	HTTPTransportConfig(fs, cfg.Transport)
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
}

func HTTPServerConfig(fs *pflag.FlagSet, cfg *forwarder.HTTPServerConfig, prefix string) {
	namePrefix := prefix
	if namePrefix != "" {
		namePrefix += "-"
	}

	usagePrefix := prefix
	if usagePrefix != "" {
		usagePrefix += " "
	}

	fs.VarP(anyflag.NewValue[forwarder.Scheme](cfg.Protocol, &cfg.Protocol,
		anyflag.EnumParser[forwarder.Scheme](forwarder.HTTPScheme, forwarder.HTTPSScheme, forwarder.HTTP2Scheme)),
		namePrefix+"protocol", "", usagePrefix+"HTTP server protocol, one of http, https, h2")
	fs.StringVarP(&cfg.Addr,
		namePrefix+"address", "", cfg.Addr, usagePrefix+"HTTP server listen address in the form of `host:port`")
	fs.StringVar(&cfg.CertFile,
		namePrefix+"cert-file", cfg.CertFile, usagePrefix+"HTTP server TLS certificate file")
	fs.StringVar(&cfg.KeyFile,
		namePrefix+"key-file", cfg.KeyFile, usagePrefix+"HTTP server TLS key file")
	fs.DurationVar(&cfg.ReadTimeout,
		namePrefix+"read-timeout", cfg.ReadTimeout, usagePrefix+"HTTP server read timeout")
	fs.VarP(anyflag.NewValue[*url.Userinfo](cfg.BasicAuth, &cfg.BasicAuth, forwarder.ParseUserInfo),
		namePrefix+"basic-auth", "", usagePrefix+"basic-auth in the form of `username:password`")
}

func LogConfig(fs *pflag.FlagSet, cfg *log.Config) {
	fs.VarP(anyflag.NewValue[*os.File](nil, &cfg.File,
		forwarder.OpenFileParser(os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600, 0o700)),
		"log-file", "", "log file path (default: stdout)")
	fs.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "enable verbose logging")
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
