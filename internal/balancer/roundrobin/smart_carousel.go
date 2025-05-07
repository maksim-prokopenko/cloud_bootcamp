package roundrobin

type infoWithWeight struct {
	info   any
	weight int
}

type SmartCarousel struct {
	unsafeCarousel *UnsafeCarousel
	checkAliveFn   CheckAliveFn
	deadUrls       map[string]infoWithWeight
}

type CheckAliveFn func(url string) bool

func NewSmartCarousel(checkAliveFn CheckAliveFn) *SmartCarousel {
	return &SmartCarousel{
		unsafeCarousel: new(UnsafeCarousel),
		checkAliveFn:   checkAliveFn,
		deadUrls:       make(map[string]infoWithWeight),
	}
}

func (cc *SmartCarousel) Set(url string, info any) {
	cc.SetWithWeight(url, info, 1)
}

func (cc *SmartCarousel) SetWithWeight(url string, info any, weight int) {
	if cc.checkAliveFn != nil && !cc.checkAliveFn(url) {
		cc.deadUrls[url] = infoWithWeight{info, weight}
		return
	}
	cc.unsafeCarousel.SetWithWeight(url, info, weight)
}

func (cc *SmartCarousel) MarkBad(url string) {
	_, ok := cc.deadUrls[url]
	if ok {
		return
	}

	info, weight, ok := cc.unsafeCarousel.ExtractWithWeight(url)
	if !ok {
		// todo what todo id url doesn't exist in simpleCarousel and in deadUrls?
	}

	cc.deadUrls[url] = infoWithWeight{info, weight}

}
