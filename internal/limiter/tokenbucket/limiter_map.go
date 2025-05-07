package tokenbucket

import (
	"golang.org/x/time/rate"
	"sync"
)

type limiterMap struct {
	mu       sync.RWMutex
	mapBuket map[string]*rate.Limiter
}

func newLimiterMap(len int) *limiterMap {
	return &limiterMap{
		mu:       sync.RWMutex{},
		mapBuket: make(map[string]*rate.Limiter, len),
	}
}

func (l *limiterMap) add(client string, limiter *rate.Limiter) {
	l.mu.Lock()
	l.mapBuket[client] = limiter
	l.mu.Unlock()
}

func (l *limiterMap) get(client string) (*rate.Limiter, bool) {
	l.mu.RLock() // TODO need only for case when add and get use same time
	lim, ok := l.mapBuket[client]
	l.mu.RUnlock()
	return lim, ok
}
