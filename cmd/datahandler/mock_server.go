package main

import (
	"net/http"
	"strconv"
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
	rus := "Очень много русского текста для корректной работы экстрактора "
	d1 := rus + "Java MySQL"
	d2 := rus + "Java Git"
	d3 := rus + "Java PHP Git"
	d4 := rus + "Go Git"
	switch url {
	case "/jobinfo/1":
		writeJson(w, []byte(`{"name": "Java Dev", "description":"`+d1+`"}`))
	case "/jobinfo/2":
		writeJson(w, []byte(`{"name": "Java Dev", "description":"`+d2+`"}`))
	case "/jobinfo/3":
		writeJson(w, []byte(`{"name": "Java Dev", "description":"`+d3+`"}`))
	case "/jobinfo/4":
		writeJson(w, []byte(`{"name": "Go Dev", "description":"`+d4+`"}`))
	default:
		writeJson(w, []byte(`{"error": "Unknown url `+url+`"}`))
	}
}

func mockServer() {
	http.HandleFunc("/pageinfo", pageHandler)
	http.HandleFunc("/jobinfo/", jobInfoHandler)
	go http.ListenAndServe(":8080", nil)
}
