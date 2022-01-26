package main

import (
	"bytes"
	"github.com/ivangodev/gonahh/entity"
	"reflect"
	"testing"
)

type JobInfo struct {
	Name     string
	Keywords []string
}

func TestCategorize(t *testing.T) {
	jobsInfo := map[string]entity.JobInfo{
		"1": {"", []string{"Java", "Git"}},
		"2": {"", []string{"C++", "MySQL"}},
	}

	buf := new(bytes.Buffer)
	buf.Write([]byte("Languages\nJava\nC++\n\nDatabases\nMySQL"))
	actual := categorize(buf, jobsInfo)
	want := map[string]string{"Java": "Languages", "C++": "Languages",
		"Git": "Other", "MySQL": "Databases"}
	if !reflect.DeepEqual(want, actual) {
		t.Fatalf("Failed categorize: actual %v VS expected %v", actual, want)
	}
}
