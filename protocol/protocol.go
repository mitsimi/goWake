package protocol

// Protocol defines the available protocols for sending a magic packet.
type Proto int

const (
	Discard Proto = iota // UDP-based Discard protocol (port 9)
	Echo                 // ICMP-based Echo protocol
)
