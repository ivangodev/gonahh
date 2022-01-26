package http

import (
	"encoding/json"
	"github.com/ivangodev/gonahh/service"
	"net/http"
)

type Delivery struct {
	service service.ServiceI
}

func NewDelivery(s service.ServiceI) *Delivery {
	return &Delivery{service: s}
}

func (d *Delivery) handleReq(w http.ResponseWriter, r *http.Request) {
	jobName := r.FormValue("jobname")
	resp, e := d.service.HandleReq(jobName)
	if e != nil {
		panic(e)
	}
	js, e := json.Marshal(resp)
	if e != nil {
		panic(e)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (d *Delivery) RegisterEndpoints() {
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		d.handleReq(w, r)
	})
	fs := http.FileServer(http.Dir("./delivery/http/static"))
	http.Handle("/", fs)
}

func (d *Delivery) Start() error {
	return http.ListenAndServe(":8080", nil)
}
