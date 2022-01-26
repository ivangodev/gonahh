package main

import (
	"bufio"
	"io"
)

type banner struct {
	banList map[string]struct{}
}

func (b *banner) banKeywords(keywords []string) (goodKeywords []string) {
	goodKeywords = []string{}
	for _, k := range keywords {
		if _, ok := b.banList[k]; !ok {
			goodKeywords = append(goodKeywords, k)
		}
	}
	return goodKeywords
}

func newKeywordsBanner(r io.Reader) (b *banner) {
	b = new(banner)
	b.banList = make(map[string]struct{})
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b.banList[scanner.Text()] = struct{}{}
	}
	return
}
