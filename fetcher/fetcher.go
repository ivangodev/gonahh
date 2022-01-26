package fetcher

import (
	"github.com/ivangodev/fefa/pkg/fefa"
	"github.com/ivangodev/gonahh/entity"
	"sync"
)

type FetchCallbacks struct {
	PageExist                func(page, area int) bool
	FetchVacanciesURLs       func(pageNb, area int) []string
	FetchVacancyDescrAndName func(url string) (descr, name string)
}

type Arear struct {
	currArea int
	Cb       FetchCallbacks
}

func NewRoot(cb FetchCallbacks) fefa.FetcherFast {
	return &Arear{Cb: cb}
}

func (a *Arear) Prepare() {
}

func (a *Arear) Next() fefa.FetcherFast {
	a.currArea++
	//Iterate over Moscow and St-Petersburg
	if a.currArea <= 2 {
		return &pagerParent{area: a.currArea, currPage: -1, cb: a.Cb}
	}
	return nil
}

func (a *Arear) CollectResults() {
}

type pagerParent struct {
	area     int
	currPage int
	cb       FetchCallbacks
}

func (p *pagerParent) Prepare() {
}

func (p *pagerParent) Next() fefa.FetcherFast {
	p.currPage++
	if p.cb.PageExist(p.currPage, p.area) {
		return &pager{area: p.area, page: p.currPage, cb: p.cb}
	}
	return nil
}

func (a *pagerParent) CollectResults() {
}

type pager struct {
	area       int
	page       int
	vacUrls    []string
	currUrlIdx int
	cb         FetchCallbacks
}

func (p *pager) Prepare() {
	p.vacUrls = p.cb.FetchVacanciesURLs(p.page, p.area)
	p.currUrlIdx = -1
}

func (p *pager) Next() fefa.FetcherFast {
	p.currUrlIdx++
	if p.currUrlIdx >= len(p.vacUrls) {
		return nil
	}
	return &vacancier{url: p.vacUrls[p.currUrlIdx], cb: p.cb}
}

func (p *pager) CollectResults() {
}

type vacancier struct {
	url string
	cb  FetchCallbacks
}

var ResVacsInfo = make(entity.URLtoDescrName)
var resMu sync.Mutex

func (v *vacancier) Prepare() {
	descr, name := v.cb.FetchVacancyDescrAndName(v.url)
	resMu.Lock()
	ResVacsInfo[v.url] = entity.DescrName{Descr: descr, Name: name}
	resMu.Unlock()
}

func (v *vacancier) Next() fefa.FetcherFast {
	return nil
}

func (v *vacancier) CollectResults() {
}
