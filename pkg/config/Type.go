package config

type Configuration struct {
	Environment Environment `json:"environment"`
}

type Environment struct {
	Target string `default:"development" json:"target"`
	Root   string `json:"root"`
}
