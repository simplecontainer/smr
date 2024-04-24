package dependency

type State struct {
	Name    string
	Success bool
}

type Result struct {
	Data string `json:"data"`
}
