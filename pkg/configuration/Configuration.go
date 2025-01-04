package configuration

func NewConfig() *Configuration {
	return &Configuration{
		Certificates: &Certificates{},
	}
}
