package extractor

import (
	"fmt"
	"github.com/grokify/html-strip-tags-go"
	"os"
	"regexp"
	"strings"
)

func isEngLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func descrInEnglish(descr string) bool {
	var runes_nb, eng_runes_nb int
	for _, rune := range descr {
		runes_nb++
		if isEngLetter(rune) {
			eng_runes_nb++
		}
	}

	/*
	 * Half of english runes is suspicous to be english-mostly text.
	 * Drop such descriptions to avoid unrelated keywords noise.
	 */
	return float64(eng_runes_nb)/float64(runes_nb) > 0.5
}

func ExtractEngWords(descrInHTML string) []string {
	descr := strip.StripTags(descrInHTML)

	if descrInEnglish(descr) {
		fmt.Fprintf(os.Stderr, "Description is likely in english %s\n", descr)
		return nil
	}

	re, _ := regexp.Compile("[a-zA-z#+]+")
	engwords := re.FindAllString(descr, -1)

	uniqueEngWords := make(map[string]interface{})
	for _, k := range engwords {
		uniqueEngWords[k] = nil
	}

	lowercaseEngWords := make(map[string]interface{})
	for k := range uniqueEngWords {
		lowercaseEngWords[strings.ToLower(k)] = nil
	}

	res := make([]string, len(lowercaseEngWords))
	i := 0
	for k := range lowercaseEngWords {
		res[i] = k
		i++
	}

	return res
}
