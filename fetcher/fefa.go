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

var fetchQueue = make(chan bool)

func fetchRateLimit(intervalMs int, reqPerIntrv int) {
	//No rate limits
	if reqPerIntrv < 0 {
		for {
			_, ok := <-fetchQueue
			if !ok {
				return
			}
		}
	}

	intrvStart := time.Now()
	currFetchersCnt := 0

	for {
		dur := time.Since(intrvStart)
		if dur.Milliseconds() > int64(intervalMs) {
			intrvStart = time.Now()
			currFetchersCnt = 0
		} else if currFetchersCnt >= reqPerIntrv {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		select {
		case _, ok := <-fetchQueue:
			if ok {
				currFetchersCnt++
			} else {
				return
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
