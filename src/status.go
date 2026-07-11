package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// CheckStatus sends an ICMP Echo Request to the target IP address purely in Go.
func CheckStatus(ip string) error {
	// targetIP may contain port (e.g., 192.168.40.51:22). Extract just the IP.
	host := ip
	if strings.Contains(ip, ":") {
		host = strings.Split(ip, ":")[0]
	}

	// 開発環境（一般ユーザー）向けに、まずUnprivileged Ping (udp4) を試み、
	// 失敗した場合や本番環境用にRawソケット (ip4:icmp) でリトライします。
	network := "udp4"
	c, err := icmp.ListenPacket(network, "0.0.0.0")
	if err != nil {
		network = "ip4:icmp"
		c, err = icmp.ListenPacket(network, "0.0.0.0")
		if err != nil {
			return fmt.Errorf("listen err: %v", err)
		}
	}
	defer c.Close()

	// Craft ICMP Echo Request
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("HELLO-WOL"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return fmt.Errorf("marshal err: %v", err)
	}

	// ネットワークタイプに応じて宛先アドレスを解決
	var dst net.Addr
	if network == "udp4" {
		dst, err = net.ResolveUDPAddr("udp4", host+":0")
	} else {
		dst, err = net.ResolveIPAddr("ip4", host)
	}
	if err != nil {
		return fmt.Errorf("resolve err: %v", err)
	}

	// Send the packet
	if _, err := c.WriteTo(wb, dst); err != nil {
		return fmt.Errorf("write err: %v", err)
	}

	// Wait for a reply with 1-second timeout
	err = c.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return fmt.Errorf("set deadline err: %v", err)
	}

	rb := make([]byte, 1500)
	n, peer, err := c.ReadFrom(rb)
	if err != nil {
		return fmt.Errorf("read err: %v", err)
	}

	rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])
	if err != nil {
		return fmt.Errorf("parse err: %v", err)
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		// udp4の場合は *net.UDPAddr, ip4の場合は *net.IPAddr が返るためIPアドレス部分だけ比較
		peerIP := ""
		switch p := peer.(type) {
		case *net.UDPAddr:
			peerIP = p.IP.String()
		case *net.IPAddr:
			peerIP = p.IP.String()
		}

		if peerIP == host {
			return nil // Got reply from the target
		}
		return fmt.Errorf("got reply from unexpected peer: %v", peerIP)
	default:
		return fmt.Errorf("got unexpected icmp message: %+v", rm)
	}
}
