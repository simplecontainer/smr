package container

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/template"
	"regexp"
	"strings"
)

// TODO: Needs refactoring

func (container *Container) Prepare(db *badger.DB, dbEncrypted *badger.DB) bool {
	var err error
	var dependencyMap []database.FormatStructure
	format := database.Format("configuration", container.Static.Group, container.Static.GeneratedName, "")

	container.Runtime.Configuration, dependencyMap, err = template.ParseTemplate(db, dbEncrypted, container.Runtime.Configuration, &format)
	container.Runtime.ObjectDependencies = append(container.Runtime.ObjectDependencies, dependencyMap...)

	if err != nil {
		return false
	}

	for keyOriginal, _ := range container.Runtime.Resources {
		container.Runtime.Resources[keyOriginal].Data, _, err = template.ParseTemplate(db, dbEncrypted, container.Runtime.Resources[keyOriginal].Data, nil)
		container.Runtime.ObjectDependencies = append(container.Runtime.ObjectDependencies, database.FormatStructure{
			Kind:       "resource",
			Group:      container.Static.Group,
			Identifier: container.Runtime.Resources[keyOriginal].Identifier,
			Key:        "",
		})
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

	for indexReadiness, _ := range container.Static.Readiness {
		for index, value := range container.Static.Readiness[indexReadiness].Body {
			regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
			matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

			if len(matches) > 0 {
				SplitByDot := strings.SplitN(matches[0][1], ".", 2)

				trimmedIndex := strings.TrimSpace(SplitByDot[1])

				if len(SplitByDot) > 1 && container.Runtime.Configuration[trimmedIndex] != nil {
					container.Static.Readiness[indexReadiness].Body[index] = strings.Replace(container.Static.Readiness[indexReadiness].Body[index], fmt.Sprintf("{{%s}}", matches[0][1]), container.Runtime.Configuration[trimmedIndex].(string), 1)
				}
			}
		}
	}

	return true
}
