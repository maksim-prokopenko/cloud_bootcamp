package carousel

import (
	"net/http"
	"sync"
)

type infoWithWeight struct {
	info   any
	weight int
}

type SmartCarousel struct {
	mu             sync.Mutex
	unsafeCarousel *UnsafeCarousel
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

func NewSmart() *SmartCarousel {
	return &SmartCarousel{
		unsafeCarousel: new(UnsafeCarousel),
		checkAliveFn:   DefaultCheckAliveFn,
		deadUrls:       make(map[string]infoWithWeight),
	}
}

func (sc *SmartCarousel) Set(url string, info any) {
	sc.SetWithWeight(url, info, 1)
}

func (sc *SmartCarousel) SetWithWeight(url string, info any, weight int) {
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
func (sc *SmartCarousel) MarkBad(url string) {
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

func (sc *SmartCarousel) GetAllUrls() []string {
	var fullList []string
	fullList = append(fullList, sc.GetHealthyUrls()...)
	fullList = append(fullList, sc.GetUnhealthyUrls()...)
	return fullList
}

func (sc *SmartCarousel) GetHealthyUrls() []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	var urls []string
	for url := range sc.unsafeCarousel.mapUrlMeta {
		urls = append(urls, url)
	}
	return urls
}

func (sc *SmartCarousel) GetUnhealthyUrls() []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	var urls []string
	for url := range sc.deadUrls {
		urls = append(urls, url)
	}
	return urls
}

func (sc *SmartCarousel) Next() any {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.unsafeCarousel.Next()
}
