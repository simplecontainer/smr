package operators

// Plugin contracts
type Operator interface {
	Run(string, ...interface{}) map[string]any
}

// Local contracts

type DependsState struct {
	Name    string
	Success bool
}
