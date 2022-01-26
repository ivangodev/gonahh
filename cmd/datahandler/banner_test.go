package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBanKeywords(t *testing.T) {
	buf := new(bytes.Buffer)
	buf.Write([]byte("developer\nsalary"))
	b := newKeywordsBanner(buf)
	actual := b.banKeywords([]string{"c++", "java", "developer", "js", "salary"})
	want := []string{"c++", "java", "js"}
	if !reflect.DeepEqual(want, actual) {
		t.Fatalf("Failed ban: actual %v VS expected %v", actual, want)
	}
}
