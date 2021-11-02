package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const apiURL = "https://api.hh.ru/vacancies"

func getVacanciesURLsPerPage(pageNb int) []string {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create new request: %v\n", err)
		os.Exit(1)
	}

	q := req.URL.Query()
	q.Add("text", "NAME:Erlang")
	q.Add("area", "2") //St-Petersburg
	q.Add("page", strconv.FormatInt(int64(pageNb), 10))
	q.Add("per_page", "10")
	req.URL.RawQuery = q.Encode()

	url := req.URL.String()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch: %v\n", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", url, err)
		os.Exit(1)
	}

	var respJSON map[string]interface{}
	if err := json.Unmarshal(body, &respJSON); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshall json %v\n", url, err)
		os.Exit(1)
	}

	res := make([]string, 0)
	for _, vacancy := range respJSON["items"].([]interface{}) {
		vacancyJSON := vacancy.(map[string]interface{})
		res = append(res, vacancyJSON["url"].(string))
	}

	return res
}

func getVacanciesURLs() []string {
	res := make([]string, 0)

	for page := 0; ; page++ {
		if v := getVacanciesURLsPerPage(page); len(v) == 0 {
			break
		} else {
			res = append(res, v...)
		}
	}

	return res
}

func main() {
	vacanciesURLs := getVacanciesURLs()
	fmt.Println(vacanciesURLs)
}
