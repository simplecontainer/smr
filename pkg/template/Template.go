package template

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/viper"
	"regexp"
	"smr/pkg/database"
	"smr/pkg/logger"
	"strings"
)

func ParseTemplate(db *badger.DB, values map[string]any) (map[string]any, error) {
	var parsedMap = make(map[string]any)
	parsedMap = values

	for keyOriginal, value := range values {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value.(string), -1)

		if len(matches) > 0 {
			for index, _ := range matches {
				SplitByDot := strings.SplitN(matches[index][1], ".", 3)

				regexExtractGroupAndId := regexp.MustCompile(`([^\[\n\]]*)`)
				GroupAndIdExtractor := regexExtractGroupAndId.FindAllStringSubmatch(SplitByDot[1], -1)

				if len(GroupAndIdExtractor) > 1 {
					format := database.Format(SplitByDot[0], GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0], SplitByDot[2])

					if format.Identifier != "*" {
						format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), GroupAndIdExtractor[1][0])
					}

					key := strings.TrimSpace(fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key))
					val, err := database.Get(db, key)

					if err != nil {
						logger.Log.Error(val)
						return nil, err
					}

					parsedMap[keyOriginal] = strings.Replace(values[keyOriginal].(string), fmt.Sprintf("{{%s}}", matches[index][1]), val, 1)
				}
			}
		}
	}

	return parsedMap, nil
}
