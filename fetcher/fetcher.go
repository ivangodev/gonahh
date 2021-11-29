package fetcher

import (
	"encoding/json"
	"fmt"
	"github.com/vanyaio/gohh/extractor"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Vacancy struct {
	URL      string
	name     string
	engWords []string
}

const apiURL = "https://api.hh.ru/vacancies"

func getVacanciesURLsPerPage(pageNb int, area string) []string {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create new request: %v\n", err)
		os.Exit(1)
	}

	q := req.URL.Query()
	q.Add("text", "NAME:developer")
	q.Add("area", area)
	q.Add("page", strconv.Itoa(pageNb))
	q.Add("per_page", "100")
	req.URL.RawQuery = q.Encode()

	url := req.URL.String()
	log.Println("Get vacancies from", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch: %s\n", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read: %s\n", err)
		os.Exit(1)
	}

	var respJSON map[string]interface{}
	if err := json.Unmarshal(body, &respJSON); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshall json: %s\n", err)
		os.Exit(1)
	}

	res := make([]string, 0)
	if respJSON["items"] == nil {
		if v, ok := respJSON["bad_argument"]; ok && v == "page, per_page" {
			return res
		} else {
			fmt.Fprintf(os.Stderr, "Unknown API message type: %v", respJSON)
			os.Exit(1)
		}
	}

	for _, vacancy := range respJSON["items"].([]interface{}) {
		vacancyJSON := vacancy.(map[string]interface{})
		res = append(res, vacancyJSON["url"].(string))
	}

	return res
}

func getVacanciesPerArea(res *[]string, area string) {
	for page := 0; ; page++ {
		if v := getVacanciesURLsPerPage(page, area); len(v) == 0 {
			break
		} else {
			*res = append(*res, v...)
		}
	}
}

func GetVacanciesURLs() []string {
	res := make([]string, 0)
	area := []string{"1", "2"}

	for _, a := range area {
		getVacanciesPerArea(&res, a)
	}

	return res
}

func fetchVacancyDescrAndName(url string) (descr string, name string) {
	client := &http.Client{}

	log.Println("Get info of vacancy", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to make new request: %s\n", err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "gonahh/1.0 (tri.ilchenko@gmail.com)")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch description: %s\n", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read: %s\n", err)
		os.Exit(1)
	}

	var respJSON map[string]interface{}
	if err := json.Unmarshal(body, &respJSON); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshall %v: %s\n", respJSON, err)
		os.Exit(1)
	}

	if respJSON["description"] == nil {
		fmt.Fprintf(os.Stderr, "Unknown API message type: %v", respJSON)
		os.Exit(1)
	}
	descr = respJSON["description"].(string)
	name = respJSON["name"].(string)
	return
}

func NewVacancies(vacanciesURLs []string) []Vacancy {
	res := make([]Vacancy, 0)

	for _, url := range vacanciesURLs {
		descr, name := fetchVacancyDescrAndName(url)
		if engwords := extractor.ExtractEngWords(descr); engwords != nil {
			vacancy := Vacancy{URL: url, name: name, engWords: engwords}
			res = append(res, vacancy)
		} else {
			fmt.Fprintf(os.Stderr, "Description of %s dropped\n", url)
		}
	}

	return res
}

func (v *Vacancy) LogVacancy(f *os.File) {
	//TODO: error handling
	f.WriteString("\n")
	f.WriteString(fmt.Sprintf("%s\n", v.URL))
	f.WriteString(fmt.Sprintf("%s\n", v.name))
	for _, w := range v.engWords {
		f.WriteString(fmt.Sprintf("%s\n", w))
	}
	f.WriteString("\n")
}
