package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestExtractEngWords(t *testing.T) {
	descr := `Много русского текста, чтобы обеспечить корректную работу. Слова:
	C++ C# Ruby javascript front-end Ruby`

	want := []string{"c++", "c#", "ruby", "javascript", "front-end"}
	actual := extractEngWords(descr)
	sort.Strings(want)
	sort.Strings(actual)
	if !reflect.DeepEqual(want, actual) {
		t.Fatalf("Failed extract: actual %v VS expected %v", actual, want)
	}
}
