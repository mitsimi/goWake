package goWake

type options struct {
	protocol Protocol
	iface    string
}

// Option is a function that modifies the options for sending a magic packet.
// It is used to configure the protocol used for sending the magic packet.
type Option func(*options)

// WithProtocol sets the protocol used for sending the magic packet.
func WithProtocol(protocol Protocol) Option {
	return func(p *options) {
		p.protocol = protocol
	}
}

// WithInterface sets the network interface used for sending the magic packet.
func WithInterface(iface string) Option {
	return func(p *options) {
		p.iface = iface
	}
}
