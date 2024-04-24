package container

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"regexp"
	"smr/pkg/database"
	"smr/pkg/logger"
	"strings"
)

func (container *Container) Prepare(db *badger.DB) bool {
	for keyOriginal, value := range container.Runtime.Configuration {
		// If {{ANYTHING_HERE}} is detected create template.Format type so that we can query the KV store if the format is valid
		format := database.FormatStructure{}

		logger.Log.Info("Trying to parse", zap.String("value", value.(string)))

		regexDetectBigBrackets := regexp.MustCompile(`{([^{{\n}}]*)}`)
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
						return false
					}

					container.Runtime.Configuration[keyOriginal] = strings.Replace(container.Runtime.Configuration[keyOriginal].(string), fmt.Sprintf("{{%s}}", matches[index][1]), val, 1)
				}
			}
		} else {
			cleanKey := strings.TrimSpace(strings.Replace(keyOriginal, fmt.Sprintf("%s.%s.%s", "configuration", container.Static.Group, container.Static.GeneratedName), "", -1))
			format = database.Format("configuration", container.Static.Group, container.Static.GeneratedName, cleanKey)

			database.Put(db, fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key), value.(string))
		}
	}

	for keyOriginal, _ := range container.Runtime.Resources {
		// If {{ ANYTHING_HERE }} is detected create template.Format type so that we can query the KV store if the format is valid
		format := database.FormatStructure{}

		for k, dataEntry := range container.Runtime.Resources[keyOriginal].Data {
			regexDetectBigBrackets := regexp.MustCompile(`{([^{{\n}}]*)}`)
			matches := regexDetectBigBrackets.FindAllStringSubmatch(dataEntry, -1)

			logger.Log.Info("Trying to parse data in the resource", zap.String("value", k))

			if len(matches) > 0 {
				for index, _ := range matches {
					SplitByDot := strings.SplitN(matches[index][1], ".", 3)

					logger.Log.Info("Detected in the resource", zap.String("value", matches[index][1]))

					regexExtractGroupAndId := regexp.MustCompile(`([^\[\n\]]*)`)
					GroupAndIdExtractor := regexExtractGroupAndId.FindAllStringSubmatch(SplitByDot[1], -1)

					if len(GroupAndIdExtractor) > 1 {
						format = database.Format(SplitByDot[0], GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0], SplitByDot[2])

						if format.Identifier != "*" {
							format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), GroupAndIdExtractor[1][0])
						}

						key := strings.TrimSpace(fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key))
						val, err := database.Get(db, key)

						if err != nil {
							logger.Log.Error(val)
							return false
						}

						logger.Log.Info("Got value from the store", zap.String("value", key))

						container.Runtime.Resources[keyOriginal].Data[k] = strings.Replace(container.Runtime.Resources[keyOriginal].Data[k], fmt.Sprintf("{{%s}}", matches[index][1]), val, 1)
					}
				}
			}
		}
	}

	for index, value := range container.Static.Env {
		regexDetectBigBrackets := regexp.MustCompile(`{([^{{}}]*)}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

		if len(matches) > 0 {
			SplitByDot := strings.SplitN(matches[0][1], ".", 2)

			trimmedIndex := strings.TrimSpace(SplitByDot[1])

			if len(SplitByDot) > 1 && container.Runtime.Configuration[trimmedIndex] != nil {
				container.Static.Env[index] = strings.Replace(container.Static.Env[index], fmt.Sprintf("{{%s}}", matches[0][1]), container.Runtime.Configuration[trimmedIndex].(string), 1)
			}
		}
	}

	return true
}
