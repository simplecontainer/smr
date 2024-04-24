package database

type FormatStructure struct {
	Kind       string
	Group      string
	Identifier string
	Key        string
}

type Response struct {
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
	Data             map[string]any
}
