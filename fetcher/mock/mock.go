package mock

import (
	"fmt"
	"github.com/ivangodev/gonahh/entity"
	"github.com/ivangodev/gonahh/fetcher"
)

const (
	pagesNumber = 2
	urlsPerPage = 3
	baseURL     = "example.com"
)

func pageExist(page, area int) bool {
	return page >= 0 && page < pagesNumber
}

func getURL(page, urlNum int) string {
	return fmt.Sprintf("%s/%d/%d", baseURL, page, urlNum)
}

func fetchVacanciesURLs(pageNb, area int) (URLs []string) {
	if !pageExist(pageNb, area) {
		return nil
	}

	URLs = make([]string, 0)
	for i := 0; i < urlsPerPage; i++ {
		URLs = append(URLs, getURL(pageNb, i))
	}
	return
}

func fetchVacancyDescrAndName(url string) (descr, name string) {
	return urlData[url].Descr, urlData[url].Name
}

var callbacks = fetcher.FetchCallbacks{
	PageExist:                pageExist,
	FetchVacanciesURLs:       fetchVacanciesURLs,
	FetchVacancyDescrAndName: fetchVacancyDescrAndName,
}

var urlData = make(entity.URLtoDescrName)

func init() {
	urlData[baseURL+"/0/0"] = entity.DescrName{Descr: "Java Python", Name: "Dev1"}
	urlData[baseURL+"/0/1"] = entity.DescrName{Descr: "Ruby", Name: "Dev2"}
	urlData[baseURL+"/0/2"] = entity.DescrName{Descr: "C++ Go Rust", Name: "Dev3"}
	urlData[baseURL+"/1/0"] = entity.DescrName{Descr: "Javascript", Name: "Dev4"}
	urlData[baseURL+"/1/1"] = entity.DescrName{Descr: "Python", Name: "Dev5"}
	urlData[baseURL+"/1/2"] = entity.DescrName{Descr: "C#", Name: "Dev6"}

	for p := 0; p < pagesNumber; p++ {
		for u := 0; u < urlsPerPage; u++ {
			url := getURL(p, u)
			_, ok := urlData[url]
			if !ok {
				panic("No data for " + url)
			}
		}
	}
}
