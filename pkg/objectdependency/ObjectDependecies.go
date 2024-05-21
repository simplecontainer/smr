package objectdependency

type DefinitionRegistry struct {
	Dependencies map[string][]string
	Order        []string
}

func NewDefinitionDependencyRegistry() *DefinitionRegistry {
	defs := DefinitionRegistry{}
	defs.Dependencies = make(map[string][]string)

	return &defs
}

func (defRegistry *DefinitionRegistry) Register(kind string, dependencies []string) {
	defRegistry.Dependencies[kind] = dependencies
}

func (defRegistry *DefinitionRegistry) GetDependencies(kind string) []string {
	dependencies, ok := defRegistry.Dependencies[kind]

	if ok {
		return dependencies
	} else {
		return []string{}
	}
}
