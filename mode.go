package forwarder

// Mode specifies mode of operation of the proxy.
type Mode uint8

//go:generate ./bin/stringer -type=Mode
const (
	Direct Mode = iota
	Upstream
	PAC
)
