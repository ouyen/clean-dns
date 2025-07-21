package main

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

type cacheEntry struct {
	msg      *dns.Msg
	expireAt time.Time
}

type dnsCache struct {
	m sync.Map
}

func (c *dnsCache) Get(key string) (*dns.Msg, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return nil, false
	}
	entry := v.(cacheEntry)
	if time.Now().After(entry.expireAt) {
		c.m.Delete(key)
		return nil, false
	}
	return entry.msg.Copy(), true
}

func (c *dnsCache) Set(key string, msg *dns.Msg, ttl uint32) {
	c.m.Store(key, cacheEntry{
		msg:      msg.Copy(),
		expireAt: time.Now().Add(time.Duration(ttl) * time.Second),
	})
}

var globalCache = &dnsCache{}
