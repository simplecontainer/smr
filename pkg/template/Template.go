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

func ParseTemplate(obj objects.ObjectInterface, retrieve map[string]string, rootFormat *f.Format) (map[string]string, []*f.Format, error) {
	var dependencyMap = make([]*f.Format, 0)
	RetrieveFromKVStore, SaveToKVStore := GetTemplatePlaceholders(retrieve, rootFormat)

	for key, placeholder := range RetrieveFromKVStore {
		formatFind := f.NewFromString(placeholder)
		keyToRetrieve := formatFind.Key

		formatFind.Key = "object"
		err := obj.Find(formatFind)

		if !obj.Exists() {
			return nil, nil, err
		}

		dependencyMap = append(dependencyMap, formatFind)

		switch formatFind.Kind {
		case "configuration":
			configuration := v1.Configuration{}

			err = json.Unmarshal(obj.GetDefinitionByte(), &configuration)

			if err != nil {
				return nil, nil, err
			}

			_, ok := configuration.Spec.Data[keyToRetrieve]

			if !ok {
				return nil, nil, errors.New(
					fmt.Sprintf("missing field in the configuration resource: %s", keyToRetrieve),
				)
			}

			RetrieveFromKVStore[key] = configuration.Spec.Data[keyToRetrieve]

			break
		}
	}

	for format, value := range SaveToKVStore {
		err := obj.Add(f.NewFromString(format), value)

		if err != nil {
			return nil, nil, err
		}
	}

	return RetrieveFromKVStore, dependencyMap, nil
}

func ParseSecretTemplate(obj objects.ObjectInterface, value string) (string, error) {
	RetrieveFromKVStore, _ := GetTemplatePlaceholders(map[string]string{
		"secret": value,
	}, nil)

	for _, placeholder := range RetrieveFromKVStore {
		format := f.NewFromString(placeholder)

		if format.Kind == "secret" {
			err := obj.Find(format)

			if !obj.Exists() {
				return value, err
			}

			value = strings.Replace(value, fmt.Sprintf("{{ %s }}", placeholder), obj.GetDefinitionString(), 1)
		}
	}

	return value, nil
}

func GetTemplatePlaceholders(values map[string]string, rootFormat *f.Format) (map[string]string, map[string]string) {
	var RetrieveFromKVStore = make(map[string]string, 0)
	var SaveToKVStore = make(map[string]string, 0)

	for keyOriginal, value := range values {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

		if len(matches) > 0 {
			for index, _ := range matches {
				format := f.NewFromString(matches[index][1])
				RetrieveFromKVStore[keyOriginal] = format.ToString()
			}
		} else {
			if rootFormat != nil {
				rootFormat.Key = keyOriginal
				SaveToKVStore[rootFormat.ToString()] = value
			}
		}
	}

	return RetrieveFromKVStore, SaveToKVStore
}
