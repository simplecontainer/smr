package relations

func NewDefinitionRelationRegistry() *RelationRegistry {
	defs := RelationRegistry{}
	defs.Relations = make(map[string][]string)

	return &defs
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
