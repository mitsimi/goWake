package goWake

import "github.com/mitsimi/goWake/v2/protocol"

type options struct {
	protocol protocol.Proto
	iface    string
}

// Option is a function that modifies the options for sending a magic packet.
// It is used to configure the protocol used for sending the magic packet.
type Option func(*options)

// WithProtocol sets the protocol used for sending the magic packet.
func WithProtocol(proto protocol.Proto) Option {
	return func(p *options) {
		p.protocol = proto
	}
}

// WithInterface sets the network interface used for sending the magic packet.
func WithInterface(iface string) Option {
	return func(p *options) {
		p.iface = iface
	}
}
