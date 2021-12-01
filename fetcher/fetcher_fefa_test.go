package fetcher

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestFetchAndLogVacs(t *testing.T) {
	filename := "example"
	fetchAndLogVacs(filename)
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
	for i := 0; i < 10; i++ {
		urls += `{"url": "http://localhost:8080/vacancy"}`
		if i != 9 {
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
	filename := "vacsFromMockServ_FeFa"
	fetchAndLogVacs(filename)
}

func TestOldFetchAndLogVacsMockServ(t *testing.T) {
	go prepareMockServ(t)
	time.Sleep(1 * time.Second)

	vacanciesURLs := GetVacanciesURLs()
	vacancies := NewVacancies(vacanciesURLs)

	filename := "vacsFromMockServ_OldFetch"
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	for _, vacancy := range vacancies {
		vacancy.LogVacancy(f)
	}
}
