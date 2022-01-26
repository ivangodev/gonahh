package mock

import (
	"github.com/ivangodev/fefa/pkg/fefa"
	"github.com/ivangodev/gonahh/fetcher"
	"reflect"
	"testing"
)

func checkResult(t *testing.T) {
	if !reflect.DeepEqual(urlData, fetcher.ResVacsInfo) {
		t.Fatalf("Unexpected result: want %v VS actual %v", urlData,
			fetcher.ResVacsInfo)
	}
}

func TestFeFa(t *testing.T) {
	fefa.FeFa(fetcher.NewRoot(callbacks), nil)
	checkResult(t)
}
