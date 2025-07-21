package main

import (
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestDNSCacheLRU(t *testing.T) {
	cache := newDNSCache(2)

	msg1 := new(dns.Msg)
	msg1.SetQuestion("a.com.", dns.TypeA)
	cache.Set("a:1", msg1, 1)

	msg2 := new(dns.Msg)
	msg2.SetQuestion("b.com.", dns.TypeA)
	cache.Set("b:1", msg2, 1)

	// Both should be present
	if _, ok := cache.Get("a:1"); !ok {
		t.Error("a:1 should be in cache")
	}
	if _, ok := cache.Get("b:1"); !ok {
		t.Error("b:1 should be in cache")
	}

	msg3 := new(dns.Msg)
	msg3.SetQuestion("c.com.", dns.TypeA)
	cache.Set("c:1", msg3, 1)

	// Now one of the old ones should be evicted (LRU)
	if _, ok := cache.Get("a:1"); ok {
		t.Error("LRU eviction failed")
	}

	// Test expiration
	time.Sleep(2 * time.Second)
	if _, ok := cache.Get("b:1"); ok {
		t.Error("b:1 should have expired")
	}
	if _, ok := cache.Get("c:1"); ok {
		t.Error("c:1 should have expired")
	}
}
