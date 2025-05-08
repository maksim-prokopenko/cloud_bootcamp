package carousel

import (
	"net/http"
	"sync"
)

type infoWithWeight struct {
	info   any
	weight int
}

type Carousel struct {
	mu             sync.Mutex
	unsafeCarousel *unsafeCarousel
	checkAliveFn   CheckAliveFn
	deadUrls       map[string]infoWithWeight
}

type CheckAliveFn func(url string) bool

var DefaultCheckAliveFn CheckAliveFn = defaultCheckAliveFn

func defaultCheckAliveFn(url string) bool {
	res, err := http.Get(url)
	if err != nil || res.StatusCode >= 500 {
		return false
	}
	return true
}

func New() *Carousel {
	return &Carousel{
		unsafeCarousel: &unsafeCarousel{
			mapUrlMeta: map[string]urlMeta{},
		},
		checkAliveFn: DefaultCheckAliveFn,
		deadUrls:     make(map[string]infoWithWeight),
	}
}

func (sc *Carousel) Set(url string, info any) {
	sc.SetWithWeight(url, info, 1)
}

func (sc *Carousel) SetWithWeight(url string, info any, weight int) {
	if sc.checkAliveFn != nil && !sc.checkAliveFn(url) {
		sc.mu.Lock()
		defer sc.mu.Unlock()
		sc.deadUrls[url] = infoWithWeight{info, weight}
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.unsafeCarousel.SetWithWeight(url, info, weight)
}

// MarkBad помечает переданный url как "нездоровый" - он исключаяется их списка выдаваемых через Next() серверов,
// но продолжает храниться с самой структуре.
// Если переданный url не был найден ни в "здоровых", ни в "нездоровых" - он просто будет проигнорирован.
func (sc *Carousel) MarkBad(url string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	_, ok := sc.deadUrls[url]
	if ok {
		return
	}
	info, weight, ok := sc.unsafeCarousel.ExtractWithWeight(url)
	if !ok {
		return
	}

	sc.deadUrls[url] = infoWithWeight{info, weight}
}

func (sc *Carousel) MarkAsGood(url string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	urlM, ok := sc.deadUrls[url]
	if !ok {
		return
	}
	delete(sc.deadUrls, url)
	sc.unsafeCarousel.SetWithWeight(url, urlM.info, urlM.weight)
}

func (sc *Carousel) GetAllUrls() (good, bad []string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	good = sc.unsafeCarousel.GetList()
	bad = make([]string, 0, len(sc.deadUrls))
	for url := range sc.deadUrls {
		bad = append(bad, url)
	}
	return good, bad
}

func (sc *Carousel) Next() any {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.unsafeCarousel.Next()
}
