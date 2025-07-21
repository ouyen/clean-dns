package main

import (
	"net"
	"testing"

	"github.com/miekg/dns"
)

func TestHandleDNSRequest(t *testing.T) {
	// Create a UDP listener for the test server
	addr := "127.0.0.1:0"
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer pc.Close()

	server := &dns.Server{PacketConn: pc, Handler: dns.HandlerFunc(handleDNSRequest)}
	go server.ActivateAndServe()
	defer server.Shutdown()

	// Build a DNS query message
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)

	c := new(dns.Client)
	resp, _, err := c.Exchange(m, pc.LocalAddr().String())
	if err != nil {
		t.Fatalf("DNS exchange failed: %v", err)
	}

	if resp == nil || resp.Rcode != dns.RcodeSuccess && resp.Rcode != dns.RcodeServerFailure {
		t.Errorf("Unexpected response: %+v", resp)
	}
}
