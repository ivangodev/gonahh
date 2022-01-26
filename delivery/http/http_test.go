package http

import (
	"encoding/json"
	"github.com/ivangodev/gonahh/entity"
	"github.com/ivangodev/gonahh/repository/mock"
	"github.com/ivangodev/gonahh/service"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHttp(t *testing.T) {
	repo := mock.NewRepoMock()
	service := service.NewService(repo)
	delivery := NewDelivery(service)
	delivery.RegisterEndpoints()

	jobNames := []string{mock.CorrectJobName, mock.CorrectJobName + "blabla"}
	for _, j := range jobNames {
		r := httptest.NewRequest(http.MethodGet, "/api/?jobname="+j, nil)
		w := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(w, r)
		res := w.Result()
		defer res.Body.Close()

		want := repo.GetServiceResp(j)
		var actual entity.ServiceResp
		if err := json.NewDecoder(res.Body).Decode(&actual); err != nil {
			t.Fatalf("Failed to decode response for %s: %s", j, err)
		}

		if !reflect.DeepEqual(want, actual) {
			t.Fatalf("Unexpected request handle for %s: want %v VS actual %v",
				j, want, actual)
		}
	}
}
