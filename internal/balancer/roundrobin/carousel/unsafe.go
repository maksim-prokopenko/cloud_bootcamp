package carousel

import "container/ring"

type UnsafeCarousel struct {
	backendRing *ring.Ring
	mapUrlMeta  map[string]urlMeta
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
	m, ok := usc.mapUrlMeta[url]
	if ok {
		// если был, обновляем его
		usc.mapUrlMeta[url] = urlMeta{
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

	urlNode := ring.New(weight)
	usc.mapUrlMeta[url] = urlMeta{
		firsRingNode: urlNode,
		weight:       weight,
		info:         info,
	}
	for i := 0; i < weight; i++ {
		urlNode.Value = url
		urlNode = urlNode.Next()
	}

	if usc.backendRing == nil {
		usc.backendRing = urlNode
	} else {
		usc.backendRing = usc.backendRing.Link(urlNode)
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

	return usc.mapUrlMeta[url].info

}
