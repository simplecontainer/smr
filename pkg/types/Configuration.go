package types

type Config struct {
	Configuration *Configuration
}

type Configuration struct {
	Environment Environment `json:"environment"`
}

type Environment struct {
	Target string `default:"development" json:"target"`
	Root   string `json:"root"`
}
