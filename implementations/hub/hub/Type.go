package hub

type Event struct {
	Kind       string
	Group      string
	Identifier string
	Data       map[string]any
}
