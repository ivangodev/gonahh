package fetcher

import (
	"sync"
	"time"
)

type fetcherFast interface {
	prepare(*GroupEndMarker)
	getGroupType() int
	next() fetcherFast
	collectResults()
}

const (
	KnownGroupSize = iota
	UnknownGroupSize
)

type GroupEndMarker struct {
	end bool
	mu  sync.Mutex
}

func (m *GroupEndMarker) markEnd() {
	m.mu.Lock()
	m.end = true
	m.mu.Unlock()
}

func (m *GroupEndMarker) isEnd() bool {
	m.mu.Lock()
	res := m.end
	m.mu.Unlock()
	return res
}

func feFa(f fetcherFast, parentGroup *sync.WaitGroup, m *GroupEndMarker) {
	f.prepare(m)
	var waitGroup sync.WaitGroup

	switch f.getGroupType() {
	case KnownGroupSize:
		for na := f.next(); na != nil; na = f.next() {
			waitGroup.Add(1)
			go feFa(na, &waitGroup, nil)
		}
	case UnknownGroupSize:
		m = new(GroupEndMarker)
		for {
			na := f.next()
			waitGroup.Add(1)
			go feFa(na, &waitGroup, m)

			if m.isEnd() {
				break
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	default:
		panic("Unknown group type")
	}

	waitGroup.Wait()
	f.collectResults()

	if parentGroup != nil {
		parentGroup.Done()
	}
}
