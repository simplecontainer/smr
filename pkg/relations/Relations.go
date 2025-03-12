package relations

import "encoding/json"

func NewDefinitionRelationRegistry() *RelationRegistry {
	defs := RelationRegistry{}
	defs.Relations = make(map[string][]string)

	return &defs
}

func (defRegistry *RelationRegistry) InTree() {
	defRegistry.Register("network", []string{""})
	defRegistry.Register("containers", []string{"network", "resource", "configuration", "certkey"})
	defRegistry.Register("gitops", []string{"certkey", "httpauth"})
	defRegistry.Register("configuration", []string{"secret"})
	defRegistry.Register("resource", []string{"configuration"})
	defRegistry.Register("certkey", []string{})
	defRegistry.Register("httpauth", []string{})
	defRegistry.Register("custom", []string{})
	defRegistry.Register("secret", []string{})
}

func (defRegistry *RelationRegistry) Register(kind string, dependencies []string) {
	defRegistry.Relations[kind] = dependencies
}

func (defRegistry *RelationRegistry) GetDependencies(kind string) []string {
	dependencies, ok := defRegistry.Relations[kind]

	if ok {
		return dependencies
	} else {
		return []string{}
	}
}

func (defRegistry *RelationRegistry) ToJson() ([]byte, error) {
	return json.Marshal(defRegistry)
}
