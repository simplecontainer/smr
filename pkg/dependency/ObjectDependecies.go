package dependency

type DefinitionRegistry struct {
	Dependencies map[string][]string
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
	dependencies, ok := defRegistry.Dependencies["foo"]

	if ok {
		return dependencies
	} else {
		return []string{}
	}
}
