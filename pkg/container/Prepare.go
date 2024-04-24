package container

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"regexp"
	"smr/pkg/template"
	"strings"
)

// TODO: Needs refactoring

func (container *Container) Prepare(db *badger.DB) bool {
	var err error
	container.Runtime.Configuration, err = template.ParseTemplate(db, container.Runtime.Configuration)

	if err != nil {
		return false
	}

	// TODO: implement saving configuration to key-value store after parsing

	for keyOriginal, _ := range container.Runtime.Resources {
		container.Runtime.Resources[keyOriginal].Data, err = template.ParseTemplate(db, container.Runtime.Resources[keyOriginal].Data)
	}

	if err != nil {
		return false
	}

	// Parse labels key not the value
	for index, _ := range container.Static.Labels {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(index, -1)

		if len(matches) > 0 {
			trimmedMatch := strings.TrimSpace(matches[0][1])
			SplitByDot := strings.SplitN(trimmedMatch, ".", 2)

			if len(SplitByDot) > 1 && container.Runtime.Configuration[SplitByDot[1]] != nil {
				newIndex := strings.Replace(index, fmt.Sprintf("{{%s}}", matches[0][1]), container.Runtime.Configuration[SplitByDot[1]].(string), 1)
				container.Static.Labels[newIndex] = container.Static.Labels[index]

				delete(container.Static.Labels, index)
			}
		}
	}

	for index, value := range container.Static.Env {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
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
