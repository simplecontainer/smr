package template

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strings"
)

func ParseTemplate(client *http.Client, values map[string]any, baseFormat *objects.FormatStructure) (map[string]any, []objects.FormatStructure, error) {
	var parsedMap = make(map[string]any)
	var dependencyMap = make([]objects.FormatStructure, 0)
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
					format := objects.Format(SplitByDot[0], GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0], SplitByDot[2])

					if format.Identifier != "*" {
						format.Identifier = fmt.Sprintf("%s-%s-%s", viper.GetString("project"), GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0])
					}

					obj := objects.New()
					err := obj.Find(client, format)

					if !obj.Exists() {
						return nil, nil, err
					}

					dependencyMap = append(dependencyMap, format)

					parsedMap[keyOriginal] = strings.Replace(values[keyOriginal].(string), fmt.Sprintf("{{%s}}", matches[index][1]), obj.GetDefinitionString(), 1)
				}
			}
		} else {
			// This is case when there is no referencing any external configuration from the container so save it in database
			if baseFormat != nil {
				baseFormat.Key = keyOriginal
				logger.Log.Info("saving into key-value store", zap.String("key", baseFormat.ToString()))

				obj := objects.New()
				obj.Add(client, *baseFormat, value.(string))
			}
		}
	}

	return parsedMap, dependencyMap, nil
}

func ParseSecretTemplate(client *http.Client, value string) (string, error) {
	regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
	matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

	if len(matches) > 0 {
		obj := objects.New()

		for index, _ := range matches {
			format := objects.FormatEmpty().FromString(matches[index][1])

			if format.Kind == "secret" {
				err := obj.Find(client, format)

				if !obj.Exists() {
					return value, err
				}
			}

			fmt.Println("SECRET")
			fmt.Println(obj)
			fmt.Println(obj.GetDefinitionString())

			value = strings.Replace(value, fmt.Sprintf("{{%s}}", matches[index][1]), obj.GetDefinitionString(), 1)
		}
	}

	return value, nil
}
