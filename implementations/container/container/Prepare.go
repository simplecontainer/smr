package container

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/template"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strings"
)

func (container *Container) Prepare(client *http.Client) bool {
	var err error
	var dependencyMap []objects.FormatStructure
	format := objects.Format("configuration", container.Static.Group, container.Static.GeneratedName, "")

	container.Runtime.Configuration, dependencyMap, err = template.ParseTemplate(client, container.Runtime.Configuration, &format)

	if err != nil {
		logger.Log.Info("container configuration parsing failed",
			zap.String("container", container.Static.GeneratedName),
			zap.String("error", err.Error()),
		)

		return false
	}

	container.Runtime.ObjectDependencies = append(container.Runtime.ObjectDependencies, dependencyMap...)

	for keyOriginal, _ := range container.Runtime.Resources {
		container.Runtime.Resources[keyOriginal].Data, _, err = template.ParseTemplate(client, container.Runtime.Resources[keyOriginal].Data, nil)

		if err != nil {
			logger.Log.Info("container configuration parsing failed",
				zap.String("container", container.Static.GeneratedName),
				zap.String("error", err.Error()),
			)

			return false
		}

		container.Runtime.ObjectDependencies = append(container.Runtime.ObjectDependencies, objects.FormatStructure{
			Kind:       "resource",
			Group:      container.Static.Group,
			Identifier: container.Runtime.Resources[keyOriginal].Identifier,
			Key:        "",
		})
	}

	// Replace placholders from the label keys in container obj with data from
	// Configuration.Runtime which has data retrieved from the KV store
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

	// Replace placholders from the envs in container obj with data from
	// Configuration.Runtime which has data retrieved from the KV store
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

	// Replace placholders from the Readiness body in container obj with data from
	// Configuration.Runtime which has data retrieved from the KV store
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
