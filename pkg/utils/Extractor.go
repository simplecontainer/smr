package utils

import (
	"regexp"
)

func ExtractGroupAndId(combined string) (group string, id string) {
	regexExtractGroupAndId := regexp.MustCompile(`([^\[\n\]]*)`)
	GroupAndIdExtractor := regexExtractGroupAndId.FindAllStringSubmatch(combined, -1)

	if len(GroupAndIdExtractor) > 1 {
		return GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0]
	} else {
		return "", ""
	}
}
