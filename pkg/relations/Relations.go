package relations

import "encoding/json"

var emptyDependencies = []string{}

func NewDefinitionRelationRegistry() *RelationRegistry {
	return &RelationRegistry{
		Relations: make(map[string][]string),
	}
}

func (defRegistry *RelationRegistry) InTree() {
	defRegistry.Register("network", emptyDependencies)
	defRegistry.Register("containers", []string{"network", "volume", "resource", "configuration", "certkey"})
	defRegistry.Register("gitops", []string{"certkey", "httpauth"})
	defRegistry.Register("configuration", []string{"secret"})
	defRegistry.Register("resource", []string{"configuration"})
	defRegistry.Register("certkey", emptyDependencies)
	defRegistry.Register("httpauth", emptyDependencies)
	defRegistry.Register("custom", emptyDependencies)
	defRegistry.Register("secret", emptyDependencies)
	defRegistry.Register("node", emptyDependencies)
	defRegistry.Register("volume", emptyDependencies)
}

func (defRegistry *RelationRegistry) Register(kind string, dependencies []string) {
	defRegistry.Relations[kind] = dependencies
}

func (defRegistry *RelationRegistry) GetDependencies(kind string) []string {
	if dependencies, ok := defRegistry.Relations[kind]; ok {
		return dependencies
	}
	return emptyDependencies
}

func (defRegistry *RelationRegistry) ToJSON() ([]byte, error) {
	return json.Marshal(defRegistry)
}
