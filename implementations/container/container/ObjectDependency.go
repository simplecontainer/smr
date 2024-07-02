package container

func (container *Container) HasDependencyOn(kind string, group string, identifier string) bool {
	for _, format := range container.Runtime.ObjectDependencies {
		if format.Identifier == identifier && format.Group == group && format.Kind == kind {
			return true
		}
	}

	return false
}
