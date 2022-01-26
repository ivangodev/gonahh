package hh

import (
	"github.com/ivangodev/fefa/pkg/fefa"
	"github.com/ivangodev/gonahh/entity"
	"github.com/ivangodev/gonahh/fetcher"
	"net/http"
	"reflect"
	"strconv"
	"testing"
)

func writeJson(w http.ResponseWriter, js []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		writeJson(w, []byte(`{"error": "Failed to handle page query parameter"}`))
		return
	}

	area, err := strconv.Atoi(r.FormValue("area"))
	if err != nil {
		writeJson(w, []byte(`{"error": "Failed to handle area query parameter"}`))
		return
	}
	if area > 1 {
		writeJson(w, []byte(`{"description": "1st area is available only"}`))
		return
	}

	switch page {
	case 0:
		writeJson(w, []byte(`{"items": [{"url": "http://localhost:8080/jobinfo/1"},
		{"url": "http://localhost:8080/jobinfo/2"}]}`))
	case 1:
		writeJson(w, []byte(`{"items": [{"url": "http://localhost:8080/jobinfo/3"},
		{"url": "http://localhost:8080/jobinfo/4"}]}`))
	default:
		writeJson(w, []byte(`{"description": "Only 2 pages available"}`))
	}
}

func jobInfoHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	switch url {
	case "/jobinfo/1":
		writeJson(w, []byte(`{"name": "Java Dev", "description": "Blabla1"}`))
	case "/jobinfo/2":
		writeJson(w, []byte(`{"name": "C++ Dev", "description": "Blabla2"}`))
	case "/jobinfo/3":
		writeJson(w, []byte(`{"name": "Ruby Dev", "description": "Blabla3"}`))
	case "/jobinfo/4":
		writeJson(w, []byte(`{"name": "JS Dev", "description": "Blabla4"}`))
	default:
		writeJson(w, []byte(`{"error": "Unknown url `+url+`"}`))
	}
}

func mockServer() {
	http.HandleFunc("/pageinfo", pageHandler)
	http.HandleFunc("/jobinfo/", jobInfoHandler)
	go http.ListenAndServe(":8080", nil)
}

func checkResult(t *testing.T) {
	want := entity.URLtoDescrName{
		"http://localhost:8080/jobinfo/1": {"Blabla1", "Java Dev"},
		"http://localhost:8080/jobinfo/2": {"Blabla2", "C++ Dev"},
		"http://localhost:8080/jobinfo/3": {"Blabla3", "Ruby Dev"},
		"http://localhost:8080/jobinfo/4": {"Blabla4", "JS Dev"},
	}
	if !reflect.DeepEqual(want, fetcher.ResVacsInfo) {
		t.Fatalf("Unexpected result: want %v VS actual %v", want,
			fetcher.ResVacsInfo)
	}
}

func TestFeFa(t *testing.T) {
	ApiUrl = "http://localhost:8080/pageinfo"
	mockServer()
	fefa.FeFa(fetcher.NewRoot(Callbacks), nil)
	checkResult(t)
}
