package main

import (
	"fmt"
	"net"
)

// SendWOL broadcasts a magic packet natively
func SendWOL(macAddr, bcastAddrStr string) error {
	mac, err := net.ParseMAC(macAddr)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %v", err)
	}

	// Craft the Magic Packet
	var packet []byte
	for i := 0; i < 6; i++ {
		packet = append(packet, 0xFF)
	}
	for i := 0; i < 16; i++ {
		packet = append(packet, mac...)
	}

	// Resolve the broadcast address
	bcastAddr, err := net.ResolveUDPAddr("udp", bcastAddrStr+":9")
	if err != nil {
		return fmt.Errorf("invalid broadcast address: %v", err)
	}

	// Dial UDP
	conn, err := net.DialUDP("udp", nil, bcastAddr)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}
