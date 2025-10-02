package ocicredentials

func New() *Credentials {
	return &Credentials{
		Registry: DefaultRegistry,
	}
}

func Default(registry string) *Credentials {
	return &Credentials{
		Registry: registry,
	}
}

func (c *Credentials) Validate() error {
	if c.Registry == "" {
		return ErrRegistryRequired
	}

	if c.Username != "" && c.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (c *Credentials) Clone() *Credentials {
	return &Credentials{
		Registry: c.Registry,
		Username: c.Username,
		Password: c.Password,
	}
}

func (c *Credentials) IsAnonymous() bool {
	return c.Username == "" && c.Password == ""
}
