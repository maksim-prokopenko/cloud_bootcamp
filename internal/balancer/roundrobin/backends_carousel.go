package roundrobin

import (
	"container/ring"
	"errors"
	"net/http"
	"sync"
)

type StrUrl string

type backendsCarousel struct {
	mu          sync.Mutex
	mapUrls     map[StrUrl]urlInfo
	backendRing *ring.Ring
}

type urlInfo struct {
	firsRingNode *ring.Ring
	weight       int
	handler      http.Handler
}

func newBackendsCarousel() *backendsCarousel {
	return &backendsCarousel{
		mapUrls: make(map[StrUrl]urlInfo),
	}
}

var ErrUrlAlreadyRegister = errors.New("url already added in carouser")

// addNew return error if url exist in carousel.
func (bc *backendsCarousel) addNew(url StrUrl, handler http.Handler) error {
	return bc.addNewWithWeight(url, handler, 1)
}

func (bc *backendsCarousel) addNewWithWeight(url StrUrl, handler http.Handler, weight int) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	_, ok := bc.mapUrls[url]
	if ok {
		return ErrUrlAlreadyRegister
	}

	urlNode := ring.New(1)
	urlNode.Value = url

	bc.mapUrls[url] = urlInfo{
		firsRingNode: urlNode,
		weight:       weight,
		handler:      handler,
	}

	if bc.backendRing == nil {
		bc.backendRing = urlNode
	} else {
		bc.backendRing.Link(urlNode)
	}

	for i := 1; i < weight; i++ {
		bc.backendRing.Link(ring.New(1)).Value = url
	}
	return nil
}

func (bc *backendsCarousel) remove(url StrUrl) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	uInfo := bc.mapUrls[url]

	uInfo.firsRingNode.Unlink(uInfo.weight)
	delete(bc.mapUrls, url)
}

func (bc *backendsCarousel) get(url StrUrl) (handler http.Handler, weight int, err error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	uInfo, ok := bc.mapUrls[url]
	if !ok {
		return nil, 0, errors.New("no one backend with this url")
	}
	return uInfo.handler, uInfo.weight, nil
}

func (bc *backendsCarousel) nextHandler() http.Handler {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	url := bc.backendRing.Value.(StrUrl)
	bc.backendRing = bc.backendRing.Next()

	return bc.mapUrls[url].handler
}
