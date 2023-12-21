package akdns

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

type RecordKey struct {
	QuestionName                string
	QuestionType, QuestionClass uint16
}

type RecordValue struct {
	Records    []dns.RR
	TTL        time.Duration
	Expiration time.Time
}

type RecordCache struct {
	Cache *sync.Map
}

func (c *RecordCache) LoadRecord(key RecordKey) ([]dns.RR, bool) {
	if records, found := c.Cache.Load(key); found {
		recordValue := records.(RecordValue)

		if recordValue.Expiration.Before(time.Now()) {
			return nil, false
		}

		records := recordValue.Records

		return records, true
	}

	return nil, false
}

func (c *RecordCache) StoreRecord(key RecordKey, records []dns.RR, ttl time.Duration) {
	expiration := time.Now().Add(time.Duration(ttl))
	recordValue := RecordValue{
		Records:    records,
		TTL:        ttl,
		Expiration: expiration,
	}

	c.Cache.Store(key, recordValue)
}

type RecordCacheError string

func (e RecordCacheError) Error() string {
	return string(e)
}
