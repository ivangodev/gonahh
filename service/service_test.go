package service

import (
	"github.com/ivangodev/gonahh/repository/mock"
	"reflect"
	"testing"
)

func TestService(t *testing.T) {
	repo := mock.NewRepoMock()
	service := NewService(repo)

	jobNames := []string{mock.CorrectJobName, mock.CorrectJobName + "blabla"}
	for _, j := range jobNames {
		want := repo.GetServiceResp(j)
		actual, err := service.HandleReq(j)
		if err != nil {
			t.Fatalf("Error to handle request for %s: %s", j, err)
		}
		if !reflect.DeepEqual(want, actual) {
			t.Fatalf("Unexpected request handle for %s: want %v VS actual %v",
				j, want, actual)
		}
	}
}
