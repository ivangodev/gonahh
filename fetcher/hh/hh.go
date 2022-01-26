package hh

import (
	"encoding/json"
	"fmt"
	"github.com/ivangodev/gonahh/fetcher"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type item struct {
	URL string `json:"url"`
}

type jobsURls struct {
	Items *[]item `json:"items"`
}

type errResp struct {
	Description *string `json:"description"`
}

type jobInfo struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

var Callbacks = fetcher.FetchCallbacks{
	PageExist:                pageExist,
	FetchVacanciesURLs:       fetchVacanciesURLs,
	FetchVacancyDescrAndName: fetchVacancyDescrAndName,
}

var ApiUrl = "https://api.hh.ru/vacancies"

func requestPageInfo(page, area int) ([]byte, error) {
	url := ApiUrl
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create new request: %w", err)
	}

	q := req.URL.Query()
	q.Add("text", "NAME:developer")
	q.Add("area", strconv.Itoa(area))
	q.Add("page", strconv.Itoa(page))
	q.Add("per_page", "100")
	req.URL.RawQuery = q.Encode()

	url = req.URL.String()
	log.Println("Get info from page", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to read body: %s", err)
	}
	return body, nil
}

func pageExist(page, area int) bool {
	body, err := requestPageInfo(page, area)
	if err != nil {
		panic(fmt.Sprintf("Failed  to request page info: %s", err))
	}

	var respJSON jobsURls
	if err := json.Unmarshal(body, &respJSON); err != nil {
		panic(fmt.Sprintf("Failed to unmarshall json %s: %s", string(body),
			err))
	}
	if respJSON.Items == nil {
		var errRespJSON errResp
		if err = json.Unmarshal(body, &errRespJSON); err != nil {
			panic(fmt.Sprintf("Failed to unmarshall json %s: %s", string(body),
				err))
		}
		if errRespJSON.Description == nil {
			panic(fmt.Sprintf("Failed to recognize json type %s", string(body)))
		}
		return false
	}
	return true
}

func fetchVacanciesURLs(page, area int) []string {
	body, err := requestPageInfo(page, area)
	if err != nil {
		panic(fmt.Sprintf("Failed  to request page info: %s", err))
	}

	var respJSON jobsURls
	if err := json.Unmarshal(body, &respJSON); err != nil {
		panic(fmt.Sprintf("Failed to unmarshall json %s: %s", string(body),
			err))
	}
	if respJSON.Items == nil {
		panic(fmt.Sprintf("Failed to recognize json type %s", string(body)))
	}

	urls := make([]string, 0, len(*respJSON.Items))
	for _, item := range *respJSON.Items {
		urls = append(urls, item.URL)
	}
	return urls
}

func fetchVacancyDescrAndName(url string) (descr string, name string) {
	client := &http.Client{}

	log.Println("Get info of vacancy", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to make new request: %s\n", err))
	}

	req.Header.Set("User-Agent", "gonahh/1.0 (ivan.swdev@gmail.com)")
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch description: %s\n", err))
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		panic(fmt.Sprintf("Failed to read body: %s", err))
	}

	var respJSON jobInfo
	if err := json.Unmarshal(body, &respJSON); err != nil {
		panic(fmt.Sprintf("Failed to unmarshall json %s: %s", string(body),
			err))
	}

	if respJSON.Description == nil || respJSON.Name == nil {
		panic(fmt.Sprintf("Failed to recognize json type %s %s", string(body), url))
	}
	descr = *respJSON.Description
	name = *respJSON.Name
	return
}
