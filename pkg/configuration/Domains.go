package configuration

func NewDomains(domains []string) *Domains {
	container := &Domains{Members: []string{}}

	for _, domain := range domains {
		container.Add(domain)
	}

	return container
}

func (domains *Domains) Add(domain string) {
	for _, d := range domains.Members {
		if d == domain {
			return
		}
	}

	if domain == "" {
		return
	}

	domains.Members = append(domains.Members, domain)
}

func (domains *Domains) Remove(domain string) {
	for i, d := range domains.Members {
		if d == domain {
			domains.Members = append(domains.Members[:i], domains.Members[i+1:]...)
		}
	}
}

func (domains *Domains) ToStringSlice() []string {
	return domains.Members
}
