package gowake

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	DefaultPort      = 9                              // Default port for discard protocol
	DefaultBroadcast = []byte{0xFF, 0xFF, 0xFF, 0xFF} // 255.255.255.255
)

// Protocol defines the available protocols for sending a magic packet.
type Protocol int

const (
	Discard Protocol = iota // UDP-based Discard protocol (port 9)
	Echo                    // ICMP-based Echo protocol
)

// Send a magic packet to a given MAC address using the default protocol (Discard/UDP).
func Send(mac string) error {
	_, err := SendWithInterface(mac, "", Discard)
	return err
}

// SendWithInterface sends a magic packet to a given MAC address, with an optional network
// interface and specified protocol. If iface is empty, it uses the default broadcast address.
// Returns the echo message if the Echo protocol is used.
func SendWithInterface(mac string, iface string, protocol Protocol) ([]byte, error) {
	var localAddr net.Addr
	var broadcastAddr net.IP = DefaultBroadcast

	if iface != "" {
		ipAddr, err := ipFromInterface(iface)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to get address for interface %s", iface), err)
		}

		localAddr = &net.UDPAddr{IP: ipAddr.IP}
		broadcastAddr, err = subnetBroadcastIP(ipAddr)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to calculate broadcast address for interface %s", iface), err)
		}
	}

	switch protocol {
	case Discard:
		return nil, sendUDPDiscard(mac, broadcastAddr, localAddr)
	case Echo:
		return sendICMPEcho(mac, broadcastAddr, localAddr)
	default:
		return nil, fmt.Errorf("unsupported protocol")
	}
}

// sendUDPDiscard sends the magic packet using UDP on the discard protocol (port 9).
func sendUDPDiscard(mac string, broadcastAddr net.IP, localAddr net.Addr) error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", broadcastAddr.String(), DefaultPort))
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

// sendICMPEcho sends the magic packet using ICMP for the Echo protocol and returns the echo message.
func sendICMPEcho(mac string, broadcastAddr net.IP, localAddr net.Addr) ([]byte, error) {
	conn, err := net.DialIP("ip4:icmp", localAddr.(*net.IPAddr), &net.IPAddr{IP: broadcastAddr})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	packet, err := NewMagicPacket(mac)
	if err != nil {
		return nil, err
	}

	data, err := packet.Marshal()
	if err != nil {
		return nil, err
	}

	// Send the packet over ICMP
	if _, err := conn.Write(data); err != nil {
		return nil, err
	}

	// Wait for an echo response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply := make([]byte, 1024)
	n, err := conn.Read(reply)
	if err != nil {
		return nil, fmt.Errorf("no response received: %v", err)
	}

	if !bytes.Equal(data, reply[:n]) {
		return nil, fmt.Errorf("received response does not match the sent packet")
	}

	fmt.Printf("Echo response received for MAC %s\n", mac)
	return reply[:n], nil
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
