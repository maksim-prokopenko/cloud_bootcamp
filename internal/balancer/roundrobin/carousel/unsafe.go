package carousel

import "container/ring"

type unsafeCarousel struct {
	backendRing *ring.Ring
	mapUrlMeta  map[string]urlMeta
}

type urlMeta struct {
	firsRingNode *ring.Ring
	weight       int
	info         any
}

func (usc *unsafeCarousel) Set(url string, info any) {
	usc.SetWithWeight(url, info, 1)
}

func (usc *unsafeCarousel) SetWithWeight(url string, info any, weight int) {

	weight = max(weight, 1) //todo ???

	// проверяем, был ли уже добавлен такой url
	_, ok := usc.mapUrlMeta[url]
	if ok {
		// если был, удаляем чтобы засетить новый
		usc.Extract(url) // todo can be more optimize
	}

	urlNode := ring.New(weight)

	for i := 0; i < weight; i++ {
		urlNode.Value = url
		urlNode = urlNode.Next()
	}

	if usc.backendRing == nil {
		usc.backendRing = urlNode
	} else {
		for _, meta := range usc.mapUrlMeta {
			meta.firsRingNode.Prev().Link(urlNode)
			break
		}
		//usc.backendRing.Prev().Link(urlNode)
	}
	usc.mapUrlMeta[url] = urlMeta{
		firsRingNode: urlNode,
		weight:       weight,
		info:         info,
	}

}

func (usc *unsafeCarousel) Extract(url string) (any, bool) {
	info, _, ok := usc.ExtractWithWeight(url)
	return info, ok
}

// TODO
func (usc *unsafeCarousel) ExtractWithWeight(url string) (any, int, bool) {
	meta, ok := usc.mapUrlMeta[url]
	if !ok {
		return nil, 0, false
	}

	// Удаляем из map
	delete(usc.mapUrlMeta, url)

	// Если кольцо состоит только из этого URL, просто очищаем его
	if usc.backendRing.Len() == meta.weight {
		usc.backendRing = nil
		return meta.info, meta.weight, true
	}

	// Находим предыдущую ноду перед первой нодой нашего URL
	prevNode := meta.firsRingNode.Prev()

	// Отсоединяем все ноды нашего URL из кольца
	extracted := prevNode.Unlink(meta.weight)

	// Проверяем, была ли текущая нода backendRing в удаленном сегменте
	// Если да, перемещаем backendRing на следующую после удаленного сегмента ноду
	for node := extracted; node != nil; node = node.Next() {
		if node == usc.backendRing {
			usc.backendRing = prevNode.Next()
			break
		}
		if node.Next() == extracted { // прошли полный круг
			break
		}
	}

	return meta.info, meta.weight, true
}

func (usc *unsafeCarousel) GetList() []string {
	list := make([]string, 0, len(usc.mapUrlMeta))
	for url := range usc.mapUrlMeta {
		list = append(list, url)
	}
	return list
}

func (usc *unsafeCarousel) Next() any {

	if usc.backendRing == nil {
		return nil
	}
	url := usc.backendRing.Value.(string)
	usc.backendRing = usc.backendRing.Next()

	return usc.mapUrlMeta[url].info

}
