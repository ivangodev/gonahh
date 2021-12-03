package fetcher

import (
	"fmt"
	"github.com/vanyaio/gohh/extractor"
	"github.com/vanyaio/gohh/vacancy"
	"log"
	"strconv"
	"sync"
)

type arear struct {
	currArea int
}

func (a *arear) prepare(m *GroupEndMarker) {
}

func (a *arear) getGroupType() int {
	return KnownGroupSize
}

func (a *arear) next() fetcherFast {
	a.currArea++
	if a.currArea <= 2 {
		return &pagerParent{area: a.currArea, currPage: -1}
	}
	return nil
}

func (a *arear) collectResults() {
}

type pagerParent struct {
	area     int
	currPage int
}

func (p *pagerParent) prepare(m *GroupEndMarker) {
}

func (p *pagerParent) getGroupType() int {
	return UnknownGroupSize
}

func (p *pagerParent) next() fetcherFast {
	p.currPage++
	return &pager{area: p.area, page: p.currPage}
}

func (a *pagerParent) collectResults() {
}

type pager struct {
	area       int
	page       int
	vacUrls    []string
	currUrlIdx int
}

func (p *pager) prepare(m *GroupEndMarker) {
	if v := getVacanciesURLsPerPage(p.page, strconv.Itoa(p.area)); len(v) == 0 {
		m.markEnd()
	} else {
		p.currUrlIdx = -1
		p.vacUrls = v
	}
}

func (p *pager) getGroupType() int {
	return KnownGroupSize
}

func (p *pager) next() fetcherFast {
	p.currUrlIdx++
	if p.currUrlIdx >= len(p.vacUrls) {
		return nil
	}
	return &vacancier{url: p.vacUrls[p.currUrlIdx]}
}

func (p *pager) collectResults() {
}

type vacancier struct {
	url string
}

func (v *vacancier) prepare(m *GroupEndMarker) {
	descr, name := fetchVacancyDescrAndName(v.url)
	if engwords := extractor.ExtractEngWords(descr); engwords != nil {
		putVacInfo(&vacancy.Vacancy{Url: v.url, Name: name,
			EngWords: engwords})
	} else {
		log.Printf("Description of %s dropped\n", v.url)
	}
}

func (v *vacancier) getGroupType() int {
	return KnownGroupSize
}

func (v *vacancier) next() fetcherFast {
	return nil
}

func (v *vacancier) collectResults() {
}

var resVacsInfo []vacancy.Vacancy = make([]vacancy.Vacancy, 0)
var resMu sync.Mutex

func putVacInfo(v *vacancy.Vacancy) {
	resMu.Lock()
	resVacsInfo = append(resVacsInfo, *v)
	resMu.Unlock()
}

func doFeFa() {
	var root arear
	feFa(&root, nil, nil)
}

func logVacs(filename string) {
	err := vacancy.LogVacancies(filename, resVacsInfo)
	if err != nil {
		panic(fmt.Sprintf("Failed to log vacancies: %s", err))
	}
}

func FetchAndLogVacs(filename string) {
	doFeFa()
	logVacs(filename)
}
