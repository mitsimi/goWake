package wol

import (
	"errors"
	"fmt"
	"net"
)

var (
	Port             = 9                              // discard protocol
	DefaultBroadcast = []byte{0xFF, 0xFF, 0xFF, 0xFF} // 255.255.255.255
)

// SendMagicPacket to send a magic packet to a given mac address.
func Send(mac string) error {
	return SendWithInterface(mac, "")
}

// SendMagicPacket to send a magic packet to a given mac address, and optionally
// receives an iface to broadcast on. An iface of "" implies a nil net.UDPAddr
func SendWithInterface(mac string, iface string) error {
	var localAddr *net.UDPAddr
	var broadcastAddr net.IP = DefaultBroadcast

	if iface != "" {
		// an interface is specified and therefore can get the local and broadcast address

		ipAddr, err := ipFromInterface(iface)
		if err != nil {
			//log.Printf("ERROR: %s\n", err.Error())
			return errors.Join(fmt.Errorf("unable to get address for interface %s", iface), err)
		}

		localAddr = &net.UDPAddr{
			IP: ipAddr.IP,
		}

		broadcastAddr, err = subnetBroadcastIP(ipAddr)
		if err != nil {
			//log.Printf("ERROR: %s\n", err.Error())
			return errors.Join(fmt.Errorf("unable to calculate broadcast address for interface %s", iface), err)
		}
	}

	// The address to broadcast to is usually the default `255.255.255.255` but
	// can be overloaded by specifying an interface from which the address gets calculated.
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", broadcastAddr.String(), Port))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", localAddr, udpAddr)
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

	//fmt.Printf("Attempting to send a magic packet to MAC %s\n", mac)
	//fmt.Printf("... Broadcasting to: %s\n", broadcastAddr)
	n, err := conn.Write(data)
	if err == nil && n != 102 {
		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
	}
	if err != nil {
		return err
	}

	//fmt.Printf("Magic packet sent successfully to %s\n", mac)
	return nil
}

// ipFromInterface returns a `*net.IPNet` from a network interface name.
func ipFromInterface(name string) (*net.IPNet, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	// Get list of unicast interface addresses
	addrs, err := iface.Addrs()
	if err == nil && len(addrs) <= 0 {
		err = fmt.Errorf("no address associated with interface %s", iface.Name)
	}
	if err != nil {
		return nil, err
	}

	// Validate that one of the addrs is a valid network IP address.
	for _, addr := range addrs {
		switch ip := addr.(type) {
		case *net.IPNet:
			if !ip.IP.IsLoopback() && ip.IP.To4() != nil {
				return ip, nil
			}
		}
	}

	return nil, fmt.Errorf("no address associated with interface %s", iface.Name)
}

// subnetBroadcastIP returns a `net.IP` which represents the broadcast address of the given `*net.IPNet“
func subnetBroadcastIP(ipnet *net.IPNet) (net.IP, error) {
	byteIp := []byte(ipnet.IP)                // []byte representation of IP
	byteMask := []byte(ipnet.Mask)            // []byte representation of mask
	byteTargetIp := make([]byte, len(byteIp)) // []byte holding target IP
	for idx := range byteIp {
		// mask will give us all fixed bits of the subnet (for the given byte)
		// inverted mask will give us all moving bits of the subnet (for the given byte)
		invertedMask := byteMask[idx] ^ 0xff // inverted mask byte

		// broadcastIP = networkIP added to the inverted mask
		byteTargetIp[idx] = byteIp[idx]&byteMask[idx] | invertedMask
	}

	return net.IP(byteTargetIp), nil
}
