package fetcher

import (
	"github.com/vanyaio/gohh/vacancy"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestFetchAndLogVacs(t *testing.T) {
	filename := "example"
	FetchAndLogVacs(filename)
}

func writeJson(w http.ResponseWriter, js []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func handleGetVacsUrls(w http.ResponseWriter, r *http.Request) {
	time.Sleep(50 * time.Millisecond)
	p := r.FormValue("page")
	if p == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Empty page number")
		return
	}

	page, err := strconv.Atoi(p)
	if err != nil {
		panic(err)
	}

	if page > 5 {
		resp := []byte(`{"bad_argument": "page, per_page"}`)
		writeJson(w, resp)
		return
	}

	var urls string
	urlsNb := 100
	for i := 0; i < urlsNb; i++ {
		urls += `{"url": "http://localhost:8080/vacancy"}`
		if i != urlsNb-1 {
			urls += `,`
		}
	}
	resp := []byte(`{ "items": [` + urls + `] }`)
	writeJson(w, resp)
}

func handleGetVacInfo(w http.ResponseWriter, r *http.Request) {
	time.Sleep(50 * time.Millisecond)
	resp := []byte(`{"description": "русский текст sql",
					 "name": "go developer"}`)
	writeJson(w, resp)
}

func prepareMockServ(t *testing.T) {
	apiURL = "http://localhost:8080/vacancies"
	http.HandleFunc("/vacancies", handleGetVacsUrls)
	http.HandleFunc("/vacancy", handleGetVacInfo)
	t.Fatal(http.ListenAndServe(":8080", nil))
}

func TestFetchAndLogVacsMockServ(t *testing.T) {
	go prepareMockServ(t)
	time.Sleep(1 * time.Second)
	go FetchRateLimit(0, -1)
	filename := "vacsFromMockServ_FeFa"
	FetchAndLogVacs(filename)
	close(FetchQueue)
}

func TestOldFetchAndLogVacsMockServ(t *testing.T) {
	go prepareMockServ(t)
	time.Sleep(1 * time.Second)
	go FetchRateLimit(0, -1)

	vacanciesURLs := GetVacanciesURLs()
	vacancies := NewVacancies(vacanciesURLs)
	close(FetchQueue)

	filename := "vacsFromMockServ_Old"
	err := vacancy.LogVacancies(filename, vacancies)
	if err != nil {
		t.Fatalf("Failed to log vacancies: %s", err)
	}
}

func TestFetchAndLogVacsRateLimit(t *testing.T) {
	go FetchRateLimit(1000, 7) //https://github.com/hhru/api/issues/74
	filename := "vacsFromRealApi_FeFa"
	FetchAndLogVacs(filename)
	close(FetchQueue)
}
