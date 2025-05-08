package carousel

import (
	"sync"
)

type Carousel struct {
	mu             sync.Mutex
	unsafeCarousel *UnsafeCarousel
}

func New() *Carousel {
	return &Carousel{
		unsafeCarousel: &UnsafeCarousel{
			mapUrlMeta: make(map[string]urlMeta),
		},
	}
}

func (sc *Carousel) Set(url string, info any) {
	sc.SetWithWeight(url, info, 1)
}

func (sc *Carousel) SetWithWeight(url string, info any, weight int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.unsafeCarousel.SetWithWeight(url, info, weight)
}

func (sc *Carousel) Extract(url string) (any, bool) {
	info, _, ok := sc.ExtractN(url)
	return info, ok
}

func (sc *Carousel) ExtractN(url string) (any, int, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.unsafeCarousel.ExtractWithWeight(url)
}

func (sc *Carousel) Next() any {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.unsafeCarousel.Next()
}
