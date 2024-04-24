package config

type Configuration struct {
	Environment Environment `yaml:"environment"`
}

type Environment struct {
	Target string `default="development" yaml:"target"`
	Root   string `yaml:"root"`
}
