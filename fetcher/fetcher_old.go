package fetcher

import (
	"fmt"
	"github.com/vanyaio/gohh/extractor"
	"github.com/vanyaio/gohh/vacancy"
	"os"
)

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

func NewVacancies(vacanciesURLs []string) []vacancy.Vacancy {
	res := make([]vacancy.Vacancy, 0)

	for _, url := range vacanciesURLs {
		descr, name := fetchVacancyDescrAndName(url)
		if engwords := extractor.ExtractEngWords(descr); engwords != nil {
			vacancy := vacancy.Vacancy{Url: url, Name: name, EngWords: engwords}
			res = append(res, vacancy)
		} else {
			fmt.Fprintf(os.Stderr, "Description of %s dropped\n", url)
		}
	}

	return res
}
