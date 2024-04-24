package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"net/http"
	"smr/pkg/container"
	"smr/pkg/definitions"
	"smr/pkg/implementations"
	"smr/pkg/logger"
	"smr/pkg/operators"
	"smr/pkg/registry"
	"sort"
)

func (api *Api) Apply(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid definition sent.",
		})
	} else {
		data := make(map[string]interface{})

		err := json.Unmarshal(jsonData, &data)
		if err != nil {
			panic(err)
		}

		if data["kind"] == "definition" {
			api.containerImplentation(jsonData, c)
		} else {
			api.internalImplementation(data["kind"].(string), jsonData, c)
		}
	}
}

func (api *Api) containerImplentation(jsonData []byte, c *gin.Context) {
	definitionSent := &definitions.Definitions{}

	if err := json.Unmarshal(jsonData, &definitionSent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Definition format that is sent is not compatible with the server definition format.",
		})
	} else {
		var globalGroups []string
		var globalNames []string

		for _, definition := range definitionSent.Definition {
			name := definition.Meta.Name
			logger.Log.Info(fmt.Sprintf("trying to create container %s", name))

			plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "implementations", "container")

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": fmt.Sprintf("Implementation is not present on the server: %s", name),
					"error":   err,
				})

				return
			}

			if plugin != nil {
				logger.Log.Info(fmt.Sprintf("plugin lookup %s", cases.Title(language.English).String("Container")))
				Implementation, err := plugin.Lookup(cases.Title(language.English).String("Container"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": fmt.Sprintf("Implementation is not present on the server. Plugin lookup failed: %s", cases.Title(language.English).String("Container")),
						"error":   err,
					})

					return
				}

				pl, ok := Implementation.(implementations.Implementation)

				if !ok {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Implementation malfunctioned on the server",
						"error":   ok,
					})

					return
				}

				groups, names := pl.Implementation(api.Manager, name, definitionSent.Definition[name])

				globalGroups = append(globalGroups, groups...)
				globalNames = append(globalNames, names...)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "implementation is not present on the server",
				})

				return
			}
		}

		order := findRelations(api.Registry, globalGroups, globalNames)

		for _, container := range order {
			solved, err := operators.Ready(api.Manager, container.Static.Group, container.Static.GeneratedName, container.Static.Definition.Spec.Container.Dependencies)

			if solved {
				logger.Log.Info("Trying to run container", zap.String("group", container.Static.Group), zap.String("name", container.Static.Name))

				api.Manager.Prepare(container)
				_, err = container.Run(api.Runtime, api.Badger)

				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})

					return
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})

				return
			}
		}

		c.JSON(http.StatusOK, api.Registry.Containers)
	}
}

func (api *Api) internalImplementation(kind string, jsonData []byte, c *gin.Context) {
	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "implementations", kind)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Internal implementation is not present on the server: %s", kind),
			"error":   err,
		})

		return
	}

	if plugin != nil {
		ImplementationInternal, err := plugin.Lookup(cases.Title(language.English).String(kind))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Internal implementation is not present on the server. Plugin lookup failed: %s", cases.Title(language.English).String(kind)),
				"error":   err,
			})

			return
		}

		pl, ok := ImplementationInternal.(implementations.ImplementationInternal)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Internal implementation malfunctioned on the server",
				"error":   ok,
			})

			return
		}

		_, err = pl.ImplementationInternal(api.Manager, jsonData)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Internal configuration parsing failed",
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Configuration saved in the key/value store",
		})

		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Internal implementation is not present on the server",
		})

		return
	}
}

func findRelations(registry registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		order = append(order, registry.Containers[groups[i]][names[i]])
	}

	sort.Sort(container.ByDepenendecies(order))

	return order
}
