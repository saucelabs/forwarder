package pac

import (
	"github.com/saucelabs/customerror"
	"github.com/saucelabs/pacman"
)

var (
	ErrFailedToCreateParser  = customerror.NewFailedToError("create PAC parser")
	ErrFailedToFindParserNil = customerror.NewFailedToError("PAC parser isn't initialized")
	ErrFailedToFindURL       = customerror.NewFailedToError("find URL")
	ErrFailedToParseURI      = customerror.NewFailedToError("parse URI")
	ErrInvalidParams         = customerror.NewInvalidError("params")
)

// Type aliasing.
type PACProxy = pacman.Proxy
type PACProxies = []PACProxy

// IParser specifies what a Parser does.
type IParser interface {
	Find(url string) ([]string, error)
}

// Parser definition.
type Parser struct {
	// uri to load PAC content. Can be remote (e.g.: HTTP) or local (path to file).
	uri string

	// pacProxiesCredential maps proxies defined in the PAC file to their
	// respective credentials.
	//
	// The original proxy auto-config specification was originally drafted by
	// Netscape in 1996. The specification hasn't changed much, and is still
	// largely the same as it was originally. It's quite simple, and there's
	// no provision for hard-coded credentials.
	pacProxiesCredentials []string

	// Underlying pac implementation.
	pac *pacman.Parser
}

// Find proxy(ies) for the given `url`.
func (pP *Parser) Find(url string) (PACProxies, error) {
	if pP == nil {
		return nil, ErrFailedToFindParserNil
	}

	return pP.pac.FindProxy(url)
}

// New is the Parser factory. It's wraps PACMan.
func New(source string, proxiesURIs ...string) (*Parser, error) {
	// Instantiate underlying PAC parser implementation.
	//
	// `uri` doesn't need to be validated, this is already done by `pacman.New`.
	// Also, there's no need to wrap `err`, Pacman is powered by `customerror`.
	// Pacman is powered by Sypl, so internal logging can be enabled by setting
	// the proper env var. See Pacman documentation.
	pacParser, err := pacman.New(source, proxiesURIs...)
	if err != nil {
		return nil, err
	}

	return &Parser{
		pacProxiesCredentials: proxiesURIs,
		uri:                   source,

		pac: pacParser,
	}, nil
}
