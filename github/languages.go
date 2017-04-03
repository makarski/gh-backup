package github

import "strings"

type Languages map[string]uint64

func (l Languages) containsAny(languages ...string) bool {
	for _, language := range languages {
		for sourceLang := range l {
			if strings.ToLower(sourceLang) == strings.ToLower(language) {
				return true
			}
		}
	}

	return false
}
