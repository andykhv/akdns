package akdns

import (
	"sync"

	"github.com/miekg/dns"
)

type RecordKey struct {
	QuestionName                string
	QuestionType, QuestionClass uint16
}

type RecordCache struct {
	Cache *sync.Map
}

func (c *RecordCache) LoadRecord(key RecordKey) ([]dns.RR, bool) {
	if records, found := c.Cache.Load(key); found {
		records := records.([]dns.RR)

		return records, true
	}

	return nil, false
}

func (c *RecordCache) StoreRecord(key RecordKey, value []dns.RR) {
	c.Cache.Store(key, value)
}

type RecordCacheError string

func (e RecordCacheError) Error() string {
	return string(e)
}
