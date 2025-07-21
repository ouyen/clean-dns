package main

import (
	"container/list"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type cacheEntry struct {
	key      string
	msg      *dns.Msg
	expireAt time.Time
}

type dnsCache struct {
	capacity int
	mu       sync.Mutex
	ll       *list.List
	cache    map[string]*list.Element
}

func newDNSCache(capacity int) *dnsCache {
	return &dnsCache{
		capacity: capacity,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
	}
}

func (c *dnsCache) Get(key string) (*dns.Msg, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, ok := c.cache[key]; ok {
		entry := ele.Value.(*cacheEntry)
		if time.Now().After(entry.expireAt) {
			c.ll.Remove(ele)
			delete(c.cache, key)
			return nil, false
		}
		c.ll.MoveToFront(ele)
		return entry.msg.Copy(), true
	}
	return nil, false
}

func (c *dnsCache) Set(key string, msg *dns.Msg, ttl uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		entry := ele.Value.(*cacheEntry)
		entry.msg = msg.Copy()
		entry.expireAt = time.Now().Add(time.Duration(ttl) * time.Second)
	} else {
		entry := &cacheEntry{
			key:      key,
			msg:      msg.Copy(),
			expireAt: time.Now().Add(time.Duration(ttl) * time.Second),
		}
		ele := c.ll.PushFront(entry)
		c.cache[key] = ele
		if c.ll.Len() > c.capacity {
			oldest := c.ll.Back()
			if oldest != nil {
				c.ll.Remove(oldest)
				delete(c.cache, oldest.Value.(*cacheEntry).key)
			}
		}
	}
}

var globalCache = newDNSCache(128)
