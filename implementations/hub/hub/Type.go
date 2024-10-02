package hub

type Event struct {
	Kind  string
	Group string
	Name  string
	Data  map[string]any
}
