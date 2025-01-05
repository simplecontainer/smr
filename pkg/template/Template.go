package template

import (
	"encoding/json"
	"errors"
	"fmt"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"regexp"
	"strings"
)

func ParseTemplate(obj objects.ObjectInterface, template string, runtime map[string]string) (string, []*f.Format, error) {
	var dependencyMap = make([]*f.Format, 0)
	placeholders := GetTemplatePlaceholders(template)

	var parsed string = template

	for _, placeholder := range placeholders {
		pf := f.NewFromString(placeholder)

		switch pf.Kind {
		case "secret":
			continue
		case "container":
			stripedIndex := strings.Replace(placeholder, "container.", "", 1)
			_, ok := runtime[stripedIndex]

			if ok {
				parsed = strings.Replace(parsed, fmt.Sprintf("{{ %s }}", placeholder), runtime[stripedIndex], -1)
			} else {
				return template, nil, errors.New(fmt.Sprintf("container runtime configuration is missing: %s", placeholder))
			}
			break
		case "network":
			obj.Find(pf)

			if !obj.Exists() {
				return template, nil, errors.New(fmt.Sprintf("object doesn't exists: %s", pf.ToString()))
			}

			parsed = strings.Replace(parsed, fmt.Sprintf("{{ %s }}", placeholder), obj.GetDefinitionString(), -1)
			break
		case "configuration":
			cf := f.NewFromString(pf.ToString())
			cf.Key = "object"

			err := obj.Find(cf)

			if !obj.Exists() {
				return template, nil, errors.New(fmt.Sprintf("object doesn't exists: %s", pf.ToString()))
			}

			dependencyMap = append(dependencyMap, cf)

			configuration := v1.ConfigurationDefinition{}

			err = json.Unmarshal(obj.GetDefinitionByte(), &configuration)

			if err != nil {
				return template, nil, err
			}

			_, ok := configuration.Spec.Data[pf.Key]

			if !ok {
				return template, nil, errors.New(
					fmt.Sprintf("missing field in the configuration resource: %s", pf.Key),
				)
			}

			parsed = strings.Replace(parsed, fmt.Sprintf("{{ %s }}", placeholder), configuration.Spec.Data[pf.Key], -1)
			break
		}
	}

	return parsed, dependencyMap, nil
}

func ParseSecretTemplate(obj objects.ObjectInterface, value string) (string, error) {
	placeholders := GetTemplatePlaceholders(value)

	for _, placeholder := range placeholders {
		format := f.NewFromString(placeholder)

		if format.Kind == "secret" {
			obj.Find(format)

			if !obj.Exists() {
				return value, errors.New(fmt.Sprintf("missing secret %s", placeholder))
			}

			value = strings.Replace(value, fmt.Sprintf("{{ %s }}", placeholder), obj.GetDefinitionString(), 1)
		}
	}

	return value, nil
}

func GetTemplatePlaceholders(template string) []string {
	var placeholders = make([]string, 0)

	regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
	matches := regexDetectBigBrackets.FindAllStringSubmatch(template, -1)

	if len(matches) > 0 {
		for index, _ := range matches {
			format := f.NewFromString(matches[index][1])
			placeholders = append(placeholders, format.ToString())
		}
	}

	return placeholders
}
