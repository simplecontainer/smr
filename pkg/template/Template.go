package template

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

func ParseTemplate(db *badger.DB, values map[string]any, baseFormat *database.FormatStructure) (map[string]any, []database.FormatStructure, error) {
	var parsedMap = make(map[string]any)
	var dependencyMap = make([]database.FormatStructure, 0)
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
						format.Identifier = fmt.Sprintf("%s-%s-%s", viper.GetString("project"), GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0])
					}

					key := strings.TrimSpace(fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key))
					val, err := database.Get(db, key)

					if err != nil {
						logger.Log.Error(err.Error(), zap.String("key", key))
						return nil, nil, err
					}

					dependencyMap = append(dependencyMap, format)
					parsedMap[keyOriginal] = strings.Replace(values[keyOriginal].(string), fmt.Sprintf("{{%s}}", matches[index][1]), val, 1)
				}
			}
		} else {
			// This is case when there is no referencing any external configuration from the container so save it in database
			if baseFormat != nil {
				baseFormat.Key = keyOriginal
				logger.Log.Info("saving into key-value store", zap.String("key", baseFormat.ToString()))
				database.Put(db, baseFormat.ToString(), value.(string))
			}
		}
	}

	return parsedMap, dependencyMap, nil
}
