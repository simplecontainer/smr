package domains

import (
	"fmt"
	"strings"
)

func New(elements ...string) Domain {
	builder := ""

	for _, member := range elements {
		builder += member
		builder += "."
	}

	builder = strings.TrimSuffix(builder, ".")
	return NewFromString(builder)
}

func NewFromString(data string) Domain {
	elements, _ := BuildElements(strings.SplitN(data, ".", 3))

	domain := Domain{
		Network:    strings.TrimSpace(elements[0]),
		Identifier: strings.TrimSpace(elements[1]),
		TLD:        strings.TrimSpace(elements[2]),
	}

	if domain.IsValid() {
		return domain
	} else {
		return Domain{}
	}
}

func BuildElements(splitted []string) ([]string, int) {
	elements := make([]string, 3)

	lengthSplitted := len(splitted)
	nonEmptyCount := 0

	for k, _ := range elements {
		if k < lengthSplitted {
			elements[k] = splitted[k]

			if splitted[k] != "" {
				nonEmptyCount += 1
			}
		} else {
			elements[k] = ""
		}
	}

	return elements, nonEmptyCount
}

func (domain Domain) IsValid() bool {
	split := strings.SplitN(domain.ToString(), ".", 3)

	for _, element := range split {
		if element == "" {
			return false
		}
	}

	return true
}

func (domain Domain) ToString() string {
	return fmt.Sprintf("%s.%s.%s", domain.Network, domain.Identifier, domain.TLD)
}

func (domain Domain) ToHeadless() string {
	return fmt.Sprintf("%s.%s.%s", domain.Network, domain.Identifier, domain.TLD)
}
