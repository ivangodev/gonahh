package main

import (
	"bufio"
	"github.com/ivangodev/gonahh/entity"
	"io"
)

const (
	pickingCategory = iota
	pickingKeyword
)

func categorize(r io.Reader, URLtoJobInfo map[string]entity.JobInfo) (keywordCategory map[string]string) {
	keywordCategory = make(map[string]string)
	scanner := bufio.NewScanner(r)
	state := pickingCategory
	var currCategory string
	for scanner.Scan() {
		str := scanner.Text()
		switch {
		case str == "":
			state = pickingCategory
		case state == pickingCategory:
			currCategory = str
			state = pickingKeyword
		case state == pickingKeyword:
			keywordCategory[str] = currCategory
		}
	}

	for _, jobInfo := range URLtoJobInfo {
		for _, keyword := range jobInfo.Keywords {
			if _, ok := keywordCategory[keyword]; !ok {
				keywordCategory[keyword] = "Other"
			}
		}
	}
	return
}
