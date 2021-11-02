package extractor

import (
	"github.com/grokify/html-strip-tags-go"
	"regexp"
	"strings"
)

func ExtractKeywords(descrInHTML string) []string {
	descr := strip.StripTags(descrInHTML)

	re, _ := regexp.Compile("[a-zA-z#+]+")
	keywords := re.FindAllString(descr, -1)

	uniqueKeywords := make(map[string]interface{})
	for _, k := range keywords {
		uniqueKeywords[k] = nil
	}

	lowercaseKeywords := make(map[string]interface{})
	for k, _ := range uniqueKeywords {
		lowercaseKeywords[strings.ToLower(k)] = nil
	}

	res := make([]string, len(lowercaseKeywords))
	i := 0
	for k := range lowercaseKeywords {
		res[i] = k
		i++
	}

	return res
}
