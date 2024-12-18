package goWake

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	defaultBroadcast = []byte{0xFF, 0xFF, 0xFF, 0xFF} // 255.255.255.255
)

// Protocol defines the available protocols for sending a magic packet.
type Protocol int

const (
	Discard Protocol = iota // UDP-based Discard protocol (port 9)
	Echo                    // ICMP-based Echo protocol
)

// Wake sends a magic packet to the specified MAC address to wake up a remote host.
// It returns an error if the magic packet could not be sent.
// By default, it uses the UDP-based Discard protocol (port 9) and sends it over all interfaces.
// The protocol and network interface can be customized using the `WithProtocol` and `WithInterface` options.
// If the Echo protocol is used, it will wait for an echo response from the remote host.
func Wake(mac string, opts ...Option) error {
	opt := options{protocol: Discard, iface: ""}
	for _, o := range opts {
		o(&opt)
	}
	return wake(mac, opt)
}

func wake(mac string, opt options) error {
	var localAddr net.Addr
	var broadcastAddr net.IP = defaultBroadcast

	if iface := opt.iface; iface != "" {
		ipAddr, err := ipFromInterface(iface)
		if err != nil {
			return errors.Join(fmt.Errorf("unable to get address for interface %s", iface), err)
		}

		localAddr = &net.UDPAddr{IP: ipAddr.IP}
		broadcastAddr, err = subnetBroadcastIP(ipAddr)
		if err != nil {
			return errors.Join(fmt.Errorf("unable to calculate broadcast address for interface %s", iface), err)
		}
	}

	switch opt.protocol {
	case Discard:
		return sendUDPDiscard(mac, broadcastAddr, localAddr)
	case Echo:
		return sendICMPEcho(mac, broadcastAddr, localAddr)
	default:
		return fmt.Errorf("unsupported protocol")
	}
}

// sendUDPDiscard sends the magic packet using UDP on the discard protocol (port 9).
func sendUDPDiscard(mac string, broadcastAddr net.IP, localAddr net.Addr) error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", broadcastAddr.String(), 9))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", localAddr.(*net.UDPAddr), udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	packet, err := NewMagicPacket(mac)
	if err != nil {
		return err
	}

	data, err := packet.Marshal()
	if err != nil {
		return err
	}

	n, err := conn.Write(data)
	if err == nil && n != 102 {
		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes)", n)
	}
	return err
}

// sendICMPEcho sends the magic packet using ICMP for the Echo protocol and awaits an answer.
func sendICMPEcho(mac string, broadcastAddr net.IP, localAddr net.Addr) error {
	conn, err := net.DialIP("ip4:icmp", localAddr.(*net.IPAddr), &net.IPAddr{IP: broadcastAddr})
	if err != nil {
		return err
	}
	defer conn.Close()

	packet, err := NewMagicPacket(mac)
	if err != nil {
		return err
	}

	data, err := packet.Marshal()
	if err != nil {
		return err
	}

	// Send the packet over ICMP
	if _, err := conn.Write(data); err != nil {
		return err
	}

	// Wait for an echo response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply := make([]byte, 1024)
	n, err := conn.Read(reply)
	if err != nil {
		return fmt.Errorf("no response received: %v", err)
	}

	if !bytes.Equal(data, reply[:n]) {
		return fmt.Errorf("received response does not match the sent packet")
	}

	return nil
}

// ipFromInterface returns a `*net.IPNet` from a network interface name.
func ipFromInterface(name string) (*net.IPNet, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil || len(addrs) == 0 {
		return nil, fmt.Errorf("no address associated with interface %s", iface.Name)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet, nil
		}
	}

	return nil, fmt.Errorf("no suitable IP address found for interface %s", iface.Name)
}

// subnetBroadcastIP calculates the broadcast address of the given `*net.IPNet`.
func subnetBroadcastIP(ipnet *net.IPNet) (net.IP, error) {
	byteIp := []byte(ipnet.IP)
	byteMask := []byte(ipnet.Mask)
	broadcastIP := make([]byte, len(byteIp))

	for i := range byteIp {
		invertedMask := byteMask[i] ^ 0xff
		broadcastIP[i] = byteIp[i]&byteMask[i] | invertedMask
	}

	return broadcastIP, nil
}
