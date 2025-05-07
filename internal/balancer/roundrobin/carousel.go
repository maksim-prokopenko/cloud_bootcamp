package roundrobin

import (
	"container/ring"
	"sync"
)

type Carousel struct {
	mu             sync.Mutex
	unsafeCarousel *UnsafeCarousel
}

func NewCarousel() *Carousel {
	return &Carousel{
		unsafeCarousel: &UnsafeCarousel{
			urlMeta: make(map[string]urlMeta),
		},
	}
}

func (sc *Carousel) Set(url string, info any) {
	sc.AddN(url, info, 1)
}

func (sc *Carousel) AddN(url string, info any, weight int) {
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

/////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

type UnsafeCarousel struct {
	backendRing *ring.Ring
	urlMeta     map[string]urlMeta
}

type urlMeta struct {
	firsRingNode *ring.Ring
	weight       int
	info         any
}

func (usc *UnsafeCarousel) Set(url string, info any) {
	usc.SetWithWeight(url, info, 1)
}

func (usc *UnsafeCarousel) SetWithWeight(url string, info any, weight int) {

	weight = max(weight, 1)

	// проверяем, был ли уже добавлен такой url
	m, ok := usc.urlMeta[url]
	if ok {
		// если был, обновляем его
		usc.urlMeta[url] = urlMeta{
			firsRingNode: m.firsRingNode, // указатель на первую ноду остается старый
			weight:       weight,         // вес заменяется на новый
			info:         info,           // инфо заменяется на новое
		}
		if weight > m.weight {
			weight -= m.weight // если вес стал больше, чем был - добавим только newWeight-oldWight новых код
		} else {
			// TODO delete extra nodes - важно удалять не сначала
			return
		}
	}

	urlNode := ring.New(1)
	urlNode.Value = url
	usc.urlMeta[url] = urlMeta{
		firsRingNode: urlNode,
		weight:       weight,
		info:         info,
	}

	if usc.backendRing == nil {
		usc.backendRing = urlNode
	} else {
		usc.backendRing.Link(urlNode)
	}
	// TODO check how it work with many wetighs - can be problem with order ring node
	for i := 1; i < weight; i++ {
		usc.backendRing.Link(ring.New(1)).Value = url
	}
}

func (usc *UnsafeCarousel) Extract(url string) (any, bool) {
	info, _, ok := usc.ExtractWithWeight(url)
	return info, ok
}

// TODO
func (usc *UnsafeCarousel) ExtractWithWeight(url string) (any, int, bool) {
	return nil, 0, false
}

func (usc *UnsafeCarousel) Next() any {

	url := usc.backendRing.Value.(string)
	usc.backendRing = usc.backendRing.Next()

	return usc.urlMeta[url].info

}
