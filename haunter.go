package fscache

import (
	"time"
)

type Entry interface {
	InUse() bool
	Name() string
}

type CacheAccessor interface {
	FileSystemStater
	EnumerateEntries(enumerator func(key string, e Entry) bool)
	RemoveFile(key string)
}

type Haunter interface {
	Haunt(c CacheAccessor)
	Next() time.Duration
}

type reaperHaunterStrategy struct {
	reaper Reaper
}

type lruHaunterStrategy struct {
	haunter LRUHaunter
}

// NewLRUHaunterStrategy returns a simple scheduleHaunt which provides an implementation LRUHaunter strategy
func NewLRUHaunterStrategy(haunter LRUHaunter) Haunter {
	return &lruHaunterStrategy{
		haunter: haunter,
	}
}

func (h *lruHaunterStrategy) Haunt(c CacheAccessor) {
	for _, key := range h.haunter.Scrub(c) {
		c.RemoveFile(key)
	}

}

func (h *lruHaunterStrategy) Next() time.Duration {
	return h.haunter.Next()
}

// NewReaperHaunterStrategy returns a simple scheduleHaunt which provides an implementation Reaper strategy
func NewReaperHaunterStrategy(reaper Reaper) Haunter {
	return &reaperHaunterStrategy{
		reaper: reaper,
	}
}

func (h *reaperHaunterStrategy) Haunt(c CacheAccessor) {
	c.EnumerateEntries(func(key string, e Entry) bool {
		if e.InUse() {
			return true
		}

		fileInfo, err := c.Stat(e.Name())
		if err != nil {
			return true
		}

		lastRead, lastWrite := fileInfo.AccessTimes()

		if h.reaper.Reap(key, lastRead, lastWrite) {
			c.RemoveFile(key)
		}

		return true
	})
}

func (h *reaperHaunterStrategy) Next() time.Duration {
	return h.reaper.Next()
}
